package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"unicode"
)

type HangmanState struct {
	PlayerName         string
	Word               string
	Display            string
	Attempts           int
	Incorrect          []string
	IncorrectWords     []string
	MaxAttempts        int
	GameOver           bool
	Win                bool
	SelectedFileName   string
	CurrentDrawingStep int
}

type Classement struct {
	Joueurs map[string]int `json:"joueurs"`
}

type JoueurClassement struct {
	Nom   string `json:"nom"`
	Score int    `json:"score"`
}

const classementFileName = "classement.json"

const saveFileName = "hangman_save.json"

var hangmanState HangmanState

func main() {
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/name", nameHandler)
	http.HandleFunc("/difficulty", difficultySelectorHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/restart", restartGameHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		state, err := loadGame()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error loading game:", err)
			return
		}

		var data interface{}
		if state != nil {
			data = map[string]interface{}{
				"Continue": true,
			}
		}

		renderTemplate(w, "index.html", data, r)
	})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		_ = deleteSaveFile()
		fmt.Println("\nSave file deleted. Exiting...")
		os.Exit(0)
	}()

	staticFileServer := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticFileServer))

	port := 8080
	fmt.Printf("Server is listening on port %d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	state, err := loadGame()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error loading game:", err)
		return
	}

	if state != nil {
		playerName := r.FormValue("playerName")
		hangmanState = HangmanState{PlayerName: playerName}
		http.Redirect(w, r, "/play?file="+state.SelectedFileName+"&playerName="+url.QueryEscape(playerName), http.StatusSeeOther)
		return
	}

	renderTemplate(w, "name.html", nil, r)
}

func nameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		playerName := r.FormValue("playerName")
		hangmanState = HangmanState{PlayerName: playerName}
		http.Redirect(w, r, "/difficulty?playerName="+url.QueryEscape(playerName), http.StatusSeeOther)
		return
	}
	renderTemplate(w, "name.html", nil, r)
}

func difficultySelectorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		selectedDifficulty := r.FormValue("selectedDifficulty")
		playerName := getPlayerName()
		http.Redirect(w, r, "/play?file="+selectedDifficulty+"&playerName="+url.QueryEscape(playerName), http.StatusSeeOther)
		return
	}

	renderTemplate(w, "difficultySelector.html", nil, r)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	selectedFile := r.FormValue("file")

	state, err := loadGame()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error loading game:", err)
		return
	}

	if selectedFile != "" && state == nil {
		resetAndStartNewGame(w, r, getPlayerName(), selectedFile)
		state, err = loadGame()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error loading game:", err)
			return
		}
	}

	if r.Method == "POST" {
		guess := r.FormValue("guess")
		wordGuess := r.FormValue("wordGuess")
		makeGuess(state, guess, wordGuess)
		saveGame(state)

		if state.GameOver {
			renderTemplate(w, "endGame.html", state, r)
			if err := deleteSaveFile(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				fmt.Println("Error deleting old save file:", err)
				return
			}
			return
		}
	}

	renderTemplate(w, "hangmanGame.html", state, r)
}

func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}, r *http.Request) {
	tmpl, err := template.ParseFiles(filepath.Join("static\\templates\\", tmplName))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Printf("Error parsing %s template: %v\n", tmplName, err)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Printf("Error executing %s template: %v\n", tmplName, err)
		return
	}
}

func resetAndStartNewGame(w http.ResponseWriter, r *http.Request, playerName, selectedFile string) {
	if r.FormValue("restart") == "true" {
		err := deleteSaveFile()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error deleting old save file:", err)
			return
		}
	}

	word, err := selectRandomWord(selectedFile)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error selecting random word:", err)
		return
	}

	newGameState := newHangmanState(playerName, word, 10, selectedFile)
	newGameState.Display = strings.Repeat("_", len(word))

	err = saveGame(newGameState)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error saving game state:", err)
		return
	}
}

func deleteSaveFile() error {
	if _, err := os.Stat(saveFileName); err == nil {
		return os.Remove(saveFileName)
	} else if os.IsNotExist(err) {
		return nil
	} else {
		return err
	}
}

func restartGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("restart") == "true" {
		playerName := getPlayerName()
		err := deleteSaveFile()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			fmt.Println("Error deleting old save file:", err)
			return
		}

		http.Redirect(w, r, "/difficulty?playerName="+url.QueryEscape(playerName), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/difficulty", http.StatusSeeOther)
}

func selectRandomWord(filename string) (string, error) {
	filePath := filepath.Join(filename)

	words, err := readWordsFromFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read words: %v", err)
	}

	return words[rand.Intn(len(words))], nil
}

func readWordsFromFile(filename string) ([]string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read %s: %v", filename, err)
	}
	words := strings.Fields(string(content))
	return words, nil
}

func newHangmanState(playerName, word string, maxAttempts int, selectedFileName string) *HangmanState {
	display := strings.Repeat("_", len(word))
	return &HangmanState{
		PlayerName:         playerName,
		Word:               word,
		Display:            display,
		Attempts:           maxAttempts,
		MaxAttempts:        maxAttempts,
		GameOver:           false,
		Win:                false,
		SelectedFileName:   selectedFileName,
		CurrentDrawingStep: 0,
	}
}

func getPlayerName() string {
	return hangmanState.PlayerName
}

func makeGuess(state *HangmanState, guess string, wordGuess string) {
	if state.GameOver {
		return
	}

	guess = strings.ToLower(guess)
	wordGuess = strings.ToLower(wordGuess)

	if wordGuess == state.Word {
		state.GameOver = true
		state.Win = true
		joueurGagne(state.PlayerName)
		return
	}

	if len(guess) == 1 && unicode.IsLetter(rune(guess[0])) {
		makeLetterGuess(state, guess)
	} else {
		makeWordGuess(state, wordGuess)
	}
}

func makeLetterGuess(state *HangmanState, guess string) {
	if strings.Contains(state.Word, guess) {
		for i, char := range state.Word {
			if string(char) == guess {
				state.Display = state.Display[:i] + guess + state.Display[i+1:]
			}
		}
		if state.Display == state.Word {
			state.GameOver = true
			state.Win = true
			joueurGagne(state.PlayerName)
		}
	} else {
		if !contains(state.Incorrect, guess) {
			state.Attempts--
			state.CurrentDrawingStep++
			state.Incorrect = append(state.Incorrect, guess)
			if state.Attempts == 0 {
				state.GameOver = true
			}
		}
	}
}

func makeWordGuess(state *HangmanState, wordGuess string) {
	if !contains(state.IncorrectWords, wordGuess) {
		state.Attempts--
		state.CurrentDrawingStep++
		state.IncorrectWords = append(state.IncorrectWords, wordGuess)
	}

	if state.Attempts == 0 {
		state.GameOver = true
	}
}

func contains(slice []string, word string) bool {
	trimmedWord := strings.TrimSpace(word)
	for _, w := range slice {
		if strings.TrimSpace(w) == trimmedWord {
			return true
		}
	}
	return false
}

func saveGame(state *HangmanState) error {
	jsonData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(saveFileName, jsonData, 0644)
	return err
}

func loadGame() (*HangmanState, error) {
	jsonData, err := ioutil.ReadFile(saveFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var state HangmanState
	err = json.Unmarshal(jsonData, &state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 - Not Found", http.StatusNotFound)
	fmt.Printf("404 - Not Found: %s\n", r.URL.Path)
}

func joueurGagne(pseudo string) {
	classement, err := chargerClassement()
	if err != nil {
		return
	}

	classement.Joueurs[pseudo]++

	err = sauvegarderClassement(classement)
	if err != nil {
		return
	}
}

func chargerClassement() (*Classement, error) {
	jsonData, err := ioutil.ReadFile(classementFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return &Classement{Joueurs: make(map[string]int)}, nil
		}
		return nil, err
	}

	var classement Classement
	err = json.Unmarshal(jsonData, &classement)
	if err != nil {
		return nil, err
	}

	return &classement, nil
}

func sauvegarderClassement(classement *Classement) error {
	jsonData, err := json.MarshalIndent(classement, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(classementFileName, jsonData, 0644)
	return err
}
