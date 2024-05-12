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
	GAME_ENTLY_LOG_FILENAME = "game_ently_log.csv"
	GAME_SCORE_LOG_FILENAME = "game_score_log.csv"
	GAME_ENTLY_LOG_HEADER   = "player_id,handle_name"
	GAME_SCORE_LOG_HEADER   = "create_timestamp,player_id,score"
	GAME_RANKING_HEADER     = "rank,player_id,handle_name,score"
	CSV_HEADER_NUM          = 0
	DISPLAY_RANKING         = 10
)

type Ranking struct {
	rank       int
	playerId   string
	handleName string
	score      int
}

type StoreScores struct {
	score       int
	playerCount int
}

// コールバック用関数: プレイヤー情報を保存する
func storePlayers(row []string, m map[string]string, i int) {
	if i == CSV_HEADER_NUM {
		header := strings.Join(row, ",")
		if header != GAME_ENTLY_LOG_HEADER {
			fmt.Fprintf(os.Stderr, "[Error]: %sファイルのヘッダーが違います。 => %s \n", GAME_ENTLY_LOG_FILENAME, header)
			os.Exit(1)
		}
		return
	}
	playerId := row[0]
	handleName := row[1]

	m[playerId] = handleName
}

// コールバック用関数: スコア情報を保存する
func storeScores(row []string, m map[string]*StoreScores, i int) {
	if i == CSV_HEADER_NUM {
		header := strings.Join(row, ",")
		if header != GAME_SCORE_LOG_HEADER {
			fmt.Fprintf(os.Stderr, "[Error]: %sファイルのヘッダーが違います。 => %s \n", GAME_SCORE_LOG_FILENAME, header)
			os.Exit(1)
		}
		return
	}

	playerId := row[1]
	score, err := strconv.Atoi(row[2])
	if err != nil {
		panic(err)
	}

	if _, ok := m[playerId]; ok {
		m[playerId].score += score
		m[playerId].playerCount++
	} else {
		m[playerId] = &StoreScores{score, 1}
	}
}

// ランキングを作成する
func makeRanking(players map[string]string, scores map[string]*StoreScores) []Ranking {
	var ranking []Ranking

	for playerId, store := range scores {
		if _, ok := players[playerId]; ok {
			// ランクは初期値0で追加する
			ranking = append(ranking, Ranking{0, playerId, players[playerId], store.score})
		}
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

// CSVファイルを1行ずつ読み込み、コールバック関数を実行する
// スコア情報のログは数千万行以上に肥大化する可能性があるため、1行ずつ読み込むようにする
func readCsv[T *StoreScores | string](file string, store map[string]T, cb func([]string, map[string]T, int)) {
	fp, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error]: %sファイルが見つかりません。\n", file)
		panic(err)
	}
	defer fp.Close()

	r := csv.NewReader(fp)
	for i := 0; ; i++ {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		cb(row, store, i)
	}
}

func getArgs() (string, string) {
	args := os.Args
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "[Error]: 引数に%sファイルが指定されていません。\n", GAME_ENTLY_LOG_FILENAME)
		os.Exit(1)
	}
	if len(args) < 3 {
		fmt.Fprintf(os.Stderr, "[Error]: 引数に%sファイルが指定されていません。\n", GAME_SCORE_LOG_FILENAME)
		os.Exit(1)
	}

	gameEntlyLogFile := args[1]
	gameScoreLogFile := args[2]

	return gameEntlyLogFile, gameScoreLogFile
}

func run() {
	// 引数を取得する
	gameEntlyLogFile, gameScoreLogFile := getArgs()

	// プレイヤー情報を保存する
	players := map[string]string{}
	readCsv[string](gameEntlyLogFile, players, storePlayers)

	// スコア情報を保存する
	scores := make(map[string]*StoreScores)
	readCsv[*StoreScores](gameScoreLogFile, scores, storeScores)

	// ランキングを作成する
	ranking := makeRanking(players, scores)

	// 作成したランキングを表示する
	displayRanking(ranking)
}

func main() {
	run()
	os.Exit(0)
}
