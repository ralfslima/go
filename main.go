// Pacote
package main

// Importações
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Modelo (Struct)
type Aluno struct {
	Codigo   string  `json:"codigo"`
	Nome     string  `json:"nome"`
	Nota1    float64 `json:"nota1"`
	Nota2    float64 `json:"nota2"`
	Media    float64 `json:"media"`
	Situacao string  `json:"situacao"`
}

// Variável global do banco de dados (Substitui o antigo slice)
var db *sql.DB

// Função para inicializar o banco de dados
func initDB() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback para testes locais no VSCode
		dbURL = "postgres://postgres:senha_local@localhost:5432/banco_local?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Erro ao abrir conexão com o banco: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Erro ao conectar no banco do Render: %v", err)
	}

	fmt.Println("Conectado ao PostgreSQL com sucesso!")

	// Criar a tabela de alunos se ela não existir
	query := `
	CREATE TABLE IF NOT EXISTS alunos (
		codigo VARCHAR(36) PRIMARY KEY,
		nome VARCHAR(100) NOT NULL,
		nota1 NUMERIC(4,2) NOT NULL,
		nota2 NUMERIC(4,2) NOT NULL,
		media NUMERIC(4,2) NOT NULL,
		situacao VARCHAR(20) NOT NULL
	);`

	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("Erro ao criar tabela alunos: %v", err)
	}
}

// Função para gerar a média e a situação
func mediaSituacao(aluno *Aluno) {
	aluno.Media = (aluno.Nota1 + aluno.Nota2) / 2

	if aluno.Media >= 7 {
		aluno.Situacao = "Aprovado(a)"
	} else if aluno.Media >= 5 {
		aluno.Situacao = "Em Recuperação"
	} else {
		aluno.Situacao = "Reprovado(a)"
	}
}

// Função para retornar um Hello World!
func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	codigoUnico := uuid.New().String()
	mensagem := map[string]string{
		"codigoUnico": codigoUnico,
		"mensagem":    "Hello World!",
	}

	json.NewEncoder(w).Encode(mensagem)
}

// Função responsável pela listagem de alunos (Busca no Banco)
func listarAlunos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Query para buscar todos os alunos
	rows, err := db.Query("SELECT codigo, nome, nota1, nota2, media, situacao FROM alunos")
	if err != nil {
		http.Error(w, "Erro ao buscar alunos no banco", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Criar uma lista vazia para garantir que retorne [] em vez de null caso o banco esteja vazio
	listaAlunos := []Aluno{}

	for rows.Next() {
		var a Aluno
		err := rows.Scan(&a.Codigo, &a.Nome, &a.Nota1, &a.Nota2, &a.Media, &a.Situacao)
		if err != nil {
			http.Error(w, "Erro ao processar dados dos alunos", http.StatusInternalServerError)
			return
		}
		listaAlunos = append(listaAlunos, a)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(listaAlunos)
}

// Função responsável pelo cadastro de um aluno (Insere no Banco)
func cadastrarAluno(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var aluno Aluno
	erro := json.NewDecoder(r.Body).Decode(&aluno)
	if erro != nil {
		http.Error(w, "Falha ao decodificar o JSON", http.StatusBadRequest)
		return
	}

	aluno.Codigo = uuid.New().String()
	mediaSituacao(&aluno)

	// Inserir no PostgreSQL
	query := "INSERT INTO alunos (codigo, nome, nota1, nota2, media, situacao) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := db.Exec(query, aluno.Codigo, aluno.Nome, aluno.Nota1, aluno.Nota2, aluno.Media, aluno.Situacao)
	if err != nil {
		http.Error(w, "Erro ao salvar aluno no banco", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(aluno)
}

// Função responsável pela alteração de dados de um aluno (Atualiza no Banco)
func alterarAluno(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	codigo := r.PathValue("codigo")

	var aluno Aluno
	erro := json.NewDecoder(r.Body).Decode(&aluno)
	if erro != nil {
		http.Error(w, "Falha ao decodificar o JSON", http.StatusBadRequest)
		return
	}

	aluno.Codigo = codigo
	mediaSituacao(&aluno)

	// Executa o UPDATE no banco e verifica se alguma linha foi alterada
	query := "UPDATE alunos SET nome = $1, nota1 = $2, nota2 = $3, media = $4, situacao = $5 WHERE codigo = $6"
	resultado, err := db.Exec(query, aluno.Nome, aluno.Nota1, aluno.Nota2, aluno.Media, aluno.Situacao, codigo)
	if err != nil {
		http.Error(w, "Erro ao atualizar dados no banco", http.StatusInternalServerError)
		return
	}

	linhasAfetadas, _ := resultado.RowsAffected()
	if linhasAfetadas == 0 {
		http.Error(w, "Código não encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(aluno)
}

// Função responsável pela remoção de um aluno (Deleta no Banco)
func removerAluno(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	codigo := r.PathValue("codigo")

	// Executa o DELETE no banco
	query := "DELETE FROM alunos WHERE codigo = $1"
	resultado, err := db.Exec(query, codigo)
	if err != nil {
		http.Error(w, "Erro ao remover aluno do banco", http.StatusInternalServerError)
		return
	}

	linhasAfetadas, _ := resultado.RowsAffected()
	if linhasAfetadas == 0 {
		http.Error(w, "Código não encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Função principal
func main() {
	// 1. Inicializa o banco de dados antes de subir o servidor
	initDB()
	defer db.Close()

	// Rotas
	http.HandleFunc("GET /", helloWorld)
	http.HandleFunc("GET /alunos", listarAlunos)
	http.HandleFunc("POST /alunos", cadastrarAluno)
	http.HandleFunc("PUT /alunos/{codigo}", alterarAluno)
	http.HandleFunc("DELETE /alunos/{codigo}", removerAluno)

	// O Render define a porta dinamicamente pela variável PORT
	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8080" // Fallback local
	}

	fmt.Printf("Servidor em execução na porta: %s\n", porta)

	// Configurar servidor ouvindo na porta correta do Render
	http.ListenAndServe(":"+porta, corsMiddleware(http.DefaultServeMux))
}

// // Pacote
// package main

// // Importações
// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"

// 	"github.com/google/uuid"
// )

// // Modelo (Struct)
// type Aluno struct {
// 	Codigo   string  `json:"codigo"`
// 	Nome     string  `json:"nome"`
// 	Nota1    float64 `json:"nota1"`
// 	Nota2    float64 `json:"nota2"`
// 	Media    float64 `json:"media"`
// 	Situacao string  `json:"situacao"`
// }

// // Slice
// var alunos = []Aluno{}

// // Função para gerar a média e a situação
// func mediaSituacao(aluno *Aluno) {
// 	// Gerar média
// 	aluno.Media = (aluno.Nota1 + aluno.Nota2) / 2

// 	// Gerar situação
// 	if aluno.Media >= 7 {
// 		aluno.Situacao = "Aprovado(a)"
// 	} else if aluno.Media >= 5 {
// 		aluno.Situacao = "Em Recuperação"
// 	} else {
// 		aluno.Situacao = "Reprovado(a)"
// 	}
// }

// // Função para retornar um Hello World!
// func helloWorld(w http.ResponseWriter, r *http.Request) {
// 	//fmt.Fprintln(w, "Hello World!")

// 	// Definir o cabeçalho
// 	w.Header().Set("Content-Type", "application/json")

// 	// Definir o Status Code
// 	w.WriteHeader(http.StatusCreated)

// 	// Gerar código único
// 	codigoUnico := uuid.New().String()

// 	// Criar JSON mensagem
// 	mensagem := map[string]string{
// 		"codigoUnico": codigoUnico,
// 		"mensagem":    "Hello World!",
// 	}

// 	// Converter o map para JSON e retornar
// 	json.NewEncoder(w).Encode(mensagem)
// }

// // Função responsável pela listagem de alunos
// func listarAlunos(w http.ResponseWriter, r *http.Request) {

// 	// Definir o cabeçalho
// 	w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 	// Definir o Status Code
// 	w.WriteHeader(http.StatusOK)

// 	// Retorna uma lista contendo todos os alunos cadastrados
// 	json.NewEncoder(w).Encode(alunos)
// }

// // Função responsável pelo cadastro de um aluno
// func cadastrarAluno(w http.ResponseWriter, r *http.Request) {

// 	// Definir o cabeçalho
// 	w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 	// Objeto do tipo Aluno
// 	var aluno Aluno

// 	// Decodificar JSON recebido
// 	erro := json.NewDecoder(r.Body).Decode(&aluno)
// 	if erro != nil {
// 		http.Error(w, "Falha ao decodificar o JSON", http.StatusBadRequest)
// 		return
// 	}

// 	// Gerar o código do aluno
// 	aluno.Codigo = uuid.New().String()

// 	// Gerar a média e a situação
// 	mediaSituacao(&aluno)

// 	// Cadastrar no Slice
// 	alunos = append(alunos, aluno)

// 	// Definir o Status Code
// 	w.WriteHeader(http.StatusCreated)

// 	// Retorna o aluno cadastrado
// 	json.NewEncoder(w).Encode(aluno)
// }

// // Função responsável pela alteração de dados de um aluno
// func alterarAluno(w http.ResponseWriter, r *http.Request) {

// 	// Definir o cabeçalho
// 	w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 	// Extrair o código do aluno
// 	codigo := r.PathValue("codigo")

// 	// Laço de repetição
// 	for indice := range alunos {

// 		// Condicional para verificar o código de cada aluno
// 		if alunos[indice].Codigo == codigo {

// 			// Objeto do tipo Aluno
// 			var aluno Aluno

// 			// Decodificar JSON recebido
// 			erro := json.NewDecoder(r.Body).Decode(&aluno)
// 			if erro != nil {
// 				http.Error(w, "Falha ao decodificar o JSON", http.StatusBadRequest)
// 				return
// 			}

// 			// Disponibilizar o código do aluno
// 			aluno.Codigo = codigo

// 			// Gerar a média e a situação
// 			mediaSituacao(&aluno)

// 			// Alterar o aluno no Slice
// 			alunos[indice] = aluno

// 			// Definir o Status Code
// 			w.WriteHeader(http.StatusOK)

// 			// Retorna o aluno com os dados alterados
// 			json.NewEncoder(w).Encode(aluno)

// 			// Finalizar a ação de alteração
// 			return

// 		}

// 	}

// 	// Retorno caso o código informado não exista
// 	http.Error(w, "Código não encontrado", http.StatusNotFound)

// }

// // Função responsável pela remoção de um aluno
// func removerAluno(w http.ResponseWriter, r *http.Request) {

// 	// Definir o cabeçalho
// 	w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 	// Extrair o código do aluno
// 	codigo := r.PathValue("codigo")

// 	// Laço de repetição
// 	for indice := range alunos {

// 		// Condicional para verificar o código de cada aluno
// 		if alunos[indice].Codigo == codigo {

// 			// Remover o aluno no Slice
// 			alunos = append(alunos[:indice], alunos[indice+1:]...)

// 			// Definir o Status Code
// 			w.WriteHeader(http.StatusNoContent)

// 			// Finalizar a ação de remoção
// 			return

// 		}

// 	}

// 	// Retorno caso o código informado não exista
// 	http.Error(w, "Código não encontrado", http.StatusNotFound)

// }

// func corsMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		// Permitir requisições do frontend
// 		w.Header().Set("Access-Control-Allow-Origin", "*")

// 		// Métodos permitidos
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

// 		// Cabeçalhos permitidos
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

// 		// Tratar requisição preflight
// 		if r.Method == http.MethodOptions {
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// // Função principal
// func main() {

// 	// Rotas
// 	http.HandleFunc("GET /", helloWorld)
// 	http.HandleFunc("GET /alunos", listarAlunos)
// 	http.HandleFunc("POST /alunos", cadastrarAluno)
// 	http.HandleFunc("PUT /alunos/{codigo}", alterarAluno)
// 	http.HandleFunc("DELETE /alunos/{codigo}", removerAluno)

// 	// Retornar o funcionamento do servidor
// 	fmt.Println("Servidor em execução no endereço: http://localhost:8080")

// 	// Configurar servidor
// 	http.ListenAndServe(":8080", corsMiddleware(http.DefaultServeMux))

// }
