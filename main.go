// Pacote
package main

// Importações
import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
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

// Slice
var alunos = []Aluno{}

// Função para gerar a média e a situação
func mediaSituacao(aluno *Aluno) {
	// Gerar média
	aluno.Media = (aluno.Nota1 + aluno.Nota2) / 2

	// Gerar situação
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
	//fmt.Fprintln(w, "Hello World!")

	// Definir o cabeçalho
	w.Header().Set("Content-Type", "application/json")

	// Definir o Status Code
	w.WriteHeader(http.StatusCreated)

	// Gerar código único
	codigoUnico := uuid.New().String()

	// Criar JSON mensagem
	mensagem := map[string]string{
		"codigoUnico": codigoUnico,
		"mensagem":    "Hello World!",
	}

	// Converter o map para JSON e retornar
	json.NewEncoder(w).Encode(mensagem)
}

// Função responsável pela listagem de alunos
func listarAlunos(w http.ResponseWriter, r *http.Request) {

	// Definir o cabeçalho
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Definir o Status Code
	w.WriteHeader(http.StatusOK)

	// Retorna uma lista contendo todos os alunos cadastrados
	json.NewEncoder(w).Encode(alunos)
}

// Função responsável pelo cadastro de um aluno
func cadastrarAluno(w http.ResponseWriter, r *http.Request) {

	// Definir o cabeçalho
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Objeto do tipo Aluno
	var aluno Aluno

	// Decodificar JSON recebido
	erro := json.NewDecoder(r.Body).Decode(&aluno)
	if erro != nil {
		http.Error(w, "Falha ao decodificar o JSON", http.StatusBadRequest)
		return
	}

	// Gerar o código do aluno
	aluno.Codigo = uuid.New().String()

	// Gerar a média e a situação
	mediaSituacao(&aluno)

	// Cadastrar no Slice
	alunos = append(alunos, aluno)

	// Definir o Status Code
	w.WriteHeader(http.StatusCreated)

	// Retorna o aluno cadastrado
	json.NewEncoder(w).Encode(aluno)
}

// Função responsável pela alteração de dados de um aluno
func alterarAluno(w http.ResponseWriter, r *http.Request) {

	// Definir o cabeçalho
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Extrair o código do aluno
	codigo := r.PathValue("codigo")

	// Laço de repetição
	for indice := range alunos {

		// Condicional para verificar o código de cada aluno
		if alunos[indice].Codigo == codigo {

			// Objeto do tipo Aluno
			var aluno Aluno

			// Decodificar JSON recebido
			erro := json.NewDecoder(r.Body).Decode(&aluno)
			if erro != nil {
				http.Error(w, "Falha ao decodificar o JSON", http.StatusBadRequest)
				return
			}

			// Disponibilizar o código do aluno
			aluno.Codigo = codigo

			// Gerar a média e a situação
			mediaSituacao(&aluno)

			// Alterar o aluno no Slice
			alunos[indice] = aluno

			// Definir o Status Code
			w.WriteHeader(http.StatusOK)

			// Retorna o aluno com os dados alterados
			json.NewEncoder(w).Encode(aluno)

			// Finalizar a ação de alteração
			return

		}

	}

	// Retorno caso o código informado não exista
	http.Error(w, "Código não encontrado", http.StatusNotFound)

}

// Função responsável pela remoção de um aluno
func removerAluno(w http.ResponseWriter, r *http.Request) {

	// Definir o cabeçalho
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Extrair o código do aluno
	codigo := r.PathValue("codigo")

	// Laço de repetição
	for indice := range alunos {

		// Condicional para verificar o código de cada aluno
		if alunos[indice].Codigo == codigo {

			// Remover o aluno no Slice
			alunos = append(alunos[:indice], alunos[indice+1:]...)

			// Definir o Status Code
			w.WriteHeader(http.StatusNoContent)

			// Finalizar a ação de remoção
			return

		}

	}

	// Retorno caso o código informado não exista
	http.Error(w, "Código não encontrado", http.StatusNotFound)

}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Permitir requisições do frontend
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Métodos permitidos
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Cabeçalhos permitidos
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Tratar requisição preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Função principal
func main() {

	// Rotas
	http.HandleFunc("GET /", helloWorld)
	http.HandleFunc("GET /alunos", listarAlunos)
	http.HandleFunc("POST /alunos", cadastrarAluno)
	http.HandleFunc("PUT /alunos/{codigo}", alterarAluno)
	http.HandleFunc("DELETE /alunos/{codigo}", removerAluno)

	// Retornar o funcionamento do servidor
	fmt.Println("Servidor em execução no endereço: http://localhost:8080")

	// Configurar servidor
	http.ListenAndServe(":8080", corsMiddleware(http.DefaultServeMux))

}
