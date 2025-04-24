package util

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type EnvVar struct {
	mu   sync.Mutex
	Vars map[string]string
}

// globalEnv is the global scoped env singleton
var globalEnv *EnvVar

func NewEnv() *EnvVar {
	if globalEnv == nil {
		globalEnv = &EnvVar{
			Vars: map[string]string{},
		}
		globalEnv.Load()
	}
	return globalEnv
}

func (env *EnvVar) Load() *EnvVar {

	env.mu.Lock()
	defer env.mu.Unlock()

	env.DotEnv("")
	expectedVars := []string{
		"WEB_PORT",
		"WEB_HOST",
		"LOCAL_SECRET",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_DATASTORE",
	}
	for _, key := range expectedVars {
		env.Vars[key] = os.Getenv(key)
	}
	if env.Expect("LOCAL_SECRET") == "" {
		bytes, err := GenerateRandomSecretKey()
		if err != nil {
			panic(err)
		}
		env.Vars["LOCAL_SECRET"] = string(bytes)
	}
	return env
}

func (env *EnvVar) Get(v string) (string, bool) {
	val, ok := env.Vars[v]
	return val, ok
}
func (env *EnvVar) Expect(v string) string {
	return env.Vars[v]
}

func (env *EnvVar) DotEnv(file string) {
	var err error
	if len(file) == 0 {
		err = godotenv.Load()
	} else {
		err = godotenv.Load(file)
	}
	if err != nil {
		log.Error().Msg("Error loading .env file")
	}
}
