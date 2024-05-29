package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	GAME_ENTLY_LOG_FILENAME string = "game_ently_log.csv"
	GAME_SCORE_LOG_FILENAME string = "game_score_log.csv"
	GAME_ENTLY_LOG_HEADER   string = "player_id,handle_name"
	GAME_SCORE_LOG_HEADER   string = "create_timestamp,player_id,score"
	GAME_RANKING_HEADER     string = "rank,player_id,handle_name,score"
	CSV_HEADER_NUM          int    = 0
	DISPLAY_RANKING         int    = 10
)

type fileNameError struct {
	fileName string
}

func (e *fileNameError) Error() string {
	return fmt.Sprintf("[Error]: 引数に%sファイルが指定されていません。\n", e.fileName)
}

type fileNotFoundError struct {
	fileName string
}

func (e *fileNotFoundError) Error() string {
	return fmt.Sprintf("[Error]: %sファイルが見つかりません。\n", e.fileName)
}

type headerCheckError struct {
	fileName    string
	errorHeader string
}

func (e *headerCheckError) Error() string {
	return fmt.Sprintf("[Error]: %sファイルのヘッダーが違います。 => %s", e.fileName, e.errorHeader)
}

type Ranking struct {
	rank       int
	playerId   string
	handleName string
	score      int
}

type Store struct {
	handleName  string
	scores      []int
	sumScore    int
	maxScore    int
	playerCount int
}

func max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}

// コールバック用関数: プレイヤー情報を保存する
func storePlayerInfo(row []string, store map[string]*Store, i int) error {
	if i == CSV_HEADER_NUM {
		header := strings.Join(row, ",")
		if header != GAME_ENTLY_LOG_HEADER {
			return &headerCheckError{GAME_ENTLY_LOG_HEADER, header}
		}
		return nil
	}
	playerId := row[0]
	handleName := row[1]

	store[playerId] = &Store{handleName: handleName}

	return nil
}

// コールバック用関数: スコア情報を保存する
func storePlayerScore(row []string, store map[string]*Store, i int) error {
	if i == CSV_HEADER_NUM {
		header := strings.Join(row, ",")
		if header != GAME_SCORE_LOG_HEADER {
			return &headerCheckError{GAME_ENTLY_LOG_HEADER, header}
		}
		return nil
	}

	playerId := row[1]
	score, err := strconv.Atoi(row[2])
	if err != nil {
		return err
	}

	if _, ok := store[playerId]; ok {
		store[playerId].scores = append(store[playerId].scores, score)
		store[playerId].sumScore += score
		store[playerId].maxScore = max(store[playerId].maxScore, score)
		store[playerId].playerCount++
	}

	return nil
}

// ランキングを作成する
func makeRanking(store map[string]*Store) []Ranking {
	var ranking []Ranking

	for playerId, store := range store {
		// ランクは初期値0で追加する
		ranking = append(ranking, Ranking{0, playerId, store.handleName, store.sumScore})
	}
	// スコア順にソートする
	sort.Slice(ranking, func(i, j int) bool {
		return ranking[i].score > ranking[j].score
	})

	// 初期値0のランクをスコア順に再設定する
	rank := 0
	prevScore := 0
	for i, v := range ranking {
		if prevScore != v.score {
			prevScore = v.score
			rank++
			ranking[i].rank = rank
		} else {
			ranking[i].rank = rank
		}
	}

	return ranking
}

// ランキングを表示する
func displayRanking(ranking []Ranking) {
	for i, v := range ranking {
		if v.rank != DISPLAY_RANKING {
			if i == CSV_HEADER_NUM {
				fmt.Fprintln(os.Stdout, GAME_RANKING_HEADER)
			}
			fmt.Fprintf(os.Stdout, "%d,%s,%s,%d\n", v.rank, v.playerId, v.handleName, v.score)
		} else {
			break
		}
	}
}

// // CSVファイルを1行ずつ読み込み、コールバック関数を実行する
// // スコア情報のログは数千万行以上に肥大化する可能性があるため、1行ずつ読み込むようにする
func readCsv(file string, store map[string]*Store, cb func([]string, map[string]*Store, int) error) error {
	fp, err := os.Open(file)
	if err != nil {
		return &fileNotFoundError{file}
	}
	defer fp.Close()

	r := csv.NewReader(fp)
	for i := 0; ; i++ {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err := cb(row, store, i); err != nil {
			return err
		}
	}

	return nil
}

func verifyArgs() (string, string, error) {
	args := os.Args
	if len(args) < 2 {
		// エラーの場合、argsは空で返す。
		return "", "", &fileNameError{GAME_ENTLY_LOG_FILENAME}
	}
	if len(args) < 3 {
		// エラーの場合、argsは空で返す。
		return "", "", &fileNameError{GAME_SCORE_LOG_FILENAME}
	}

	gameEntlyLogFile := args[1]
	gameScoreLogFile := args[2]

	return gameEntlyLogFile, gameScoreLogFile, nil
}

func run() {
	// 引数を取得する
	gameEntlyLogFile, gameScoreLogFile, err := verifyArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// プレイヤー&スコア情報を保存するStoreを定義する
	store := make(map[string]*Store, 10000)

	// プレイヤー情報をStoreに保存する
	if err := readCsv(gameEntlyLogFile, store, storePlayerInfo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// スコア情報をStoreに保存する
	if err := readCsv(gameScoreLogFile, store, storePlayerScore); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// ランキングを作成する
	ranking := makeRanking(store)

	// 作成したランキングを表示する
	displayRanking(ranking)
}

func main() {
	run()
	os.Exit(0)
}
