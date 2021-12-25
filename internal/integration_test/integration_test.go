package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fahmifan/smol/internal/config"
	"github.com/fahmifan/smol/internal/model/models"
	"github.com/fahmifan/smol/internal/restapi"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

// login url: http://localhost:8000/api/auth/login/oauth2?provider=google
// run the test in single thread top down sequential
var client = http.DefaultClient
var baseURL = config.ServerBaseURL()
var integrationCfg = struct {
	UserID       string `json:"userID"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}{}

const MB = 1024 * 1024

// 2 MB
const maxBytes = 2 * MB

func init() {
	config.InitLogger()
	models.LogErr(godotenv.Load("../../.env"))

	cfgFile, err := os.Open("../../integration.cfg.json")
	models.PanicErr(err)
	cfgBt, err := io.ReadAll(cfgFile)
	models.PanicErr(err)
	unmarshal(cfgBt, &integrationCfg)

	client.Timeout = time.Second * 5
}

func unmarshal(bt []byte, i interface{}) {
	if bt == nil || i == nil {
		return
	}

	err := json.Unmarshal(bt, i)
	if err != nil {
		log.Error().Err(err).Msg("")
	}
}

func TestInstrument(t *testing.T) {
	newTodo := createTodo(t)
	newTodo2 := createTodo(t)
	todos := listTodos(t)
	mustContainsTodo(t, newTodo, todos)
	mustContainsTodo(t, newTodo2, todos)
}

func mustContainsTodo(t *testing.T, targetTodo restapi.Todo, todos []restapi.Todo) {
	ok := false
	for _, todo := range todos {
		if todo == targetTodo {
			ok = true
		}
	}
	assert.True(t, ok)
}

func createTodo(t *testing.T) restapi.Todo {
	addTodoReq := restapi.AddTodoRequest{
		Detail: `foobar`,
		Done:   true,
	}
	req, err := http.NewRequest(http.MethodPost,
		baseURL+"/api/todos",
		strings.NewReader(models.JSONS(addTodoReq)),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "aaplication/json")
	req.Header.Set("Authorization", "Bearer "+integrationCfg.AccessToken)

	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.EqualValues(t, http.StatusOK, res.StatusCode)

	bodyReader := http.MaxBytesReader(nil, res.Body, maxBytes)
	defer bodyReader.Close()

	body, err := io.ReadAll(bodyReader)
	assert.NoError(t, err)

	todoRes := restapi.Todo{}
	unmarshal(body, &todoRes)
	assert.NoError(t, err)

	assert.NotEmpty(t, todoRes.ID)
	assert.Equal(t, `foobar`, todoRes.Detail)

	return todoRes
}

func listTodos(t *testing.T) []restapi.Todo {
	req, err := http.NewRequest(http.MethodGet,
		baseURL+"/api/todos",
		strings.NewReader(`
		{
			"pagination": {
				"cursor": "",
				"backward": false,
				"size": 10
			}
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+integrationCfg.AccessToken)

	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.EqualValues(t, http.StatusOK, res.StatusCode)

	bodyReader := http.MaxBytesReader(nil, res.Body, maxBytes)
	defer bodyReader.Close()

	body, err := io.ReadAll(bodyReader)
	assert.NoError(t, err)

	var resPagin restapi.ResponseWithPagination
	unmarshal(body, &resPagin)

	var todos []restapi.Todo
	unmarshal(models.JSON(resPagin.Data), &todos)
	assert.Greater(t, len(todos), 1)

	return todos
}
