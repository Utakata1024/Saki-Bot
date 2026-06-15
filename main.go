package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	// 1. .env ファイルから環境変数を読み込む
	err := godotenv.Load()
	if err != nil {
		log.Println(".env ファイルが見つかりません。環境変数が直接設定されているとみなします。")
	}

	// 2. 環境変数からトークンを取得
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("環境変数 DISCORD_BOT_TOKEN が設定されていません")
	}

	// 3. Discordセッションを作成
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Discordセッションの作成に失敗しました: %v", err)
	}

	// ハンドラ（イベントが起きた時の処理）を登録
	dg.AddHandler(messageCreate)

	// メッセージ内容を読み取るためのインテントを設定
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	// Discordに接続
	err = dg.Open()
	if err != nil {
		log.Fatalf("Discordへの接続に失敗しました: %v", err)
	}

	fmt.Println("Botが起動しました。Ctrl+C で終了します。")

	// 終了シグナルを待機
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

// メッセージが投稿されたときに呼ばれる関数
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Bot自身の発言は無視
	if m.Author.ID == s.State.User.ID {
		return
	}

	// コマンドの判定
	if strings.HasPrefix(m.Content, "!!help") {
		// 【機能1】コマンドリストの提示
		helpMessage := "🤖 **利用可能なコマンドリスト** 🤖\n" +
			"`!!help` - このコマンドリストを表示します。\n" +
			"`!!time` - 現在の日本時間を教えます。\n" +
			"`!!remind [分] [内容]` - 指定した分後にリマインドします（例: `!!remind 3 お湯を入れた`）"
		s.ChannelMessageSend(m.ChannelID, helpMessage)
		return
	}

	if strings.HasPrefix(m.Content, "!!time") {
		// 【機能2】時間を教えてくれる機能
		// 日本時間を取得
		loc, _ := time.LoadLocation("Asia/Tokyo")
		now := time.Now().In(loc)
		timeStr := now.Format("2006年01月02日 15時04分05秒")
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⏰ 現在時刻は **%s** です。", timeStr))
		return
	}

	if strings.HasPrefix(m.Content, "!!remind ") {
		// 【機能3】リマインド機能
		parts := strings.SplitN(m.Content, " ", 3)
		if len(parts) < 3 {
			s.ChannelMessageSend(m.ChannelID, "⚠️ 使い方が間違っています。\n例: `!!remind 5 筋トレをする`")
			return
		}

		minutesStr := parts[1]
		content := parts[2]

		minutes, err := strconv.Atoi(minutesStr)
		if err != nil || minutes <= 0 {
			s.ChannelMessageSend(m.ChannelID, "⚠️ 時間は正の整数（分）で指定してください。")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ %d分後にリマインドします: 「%s」", minutes, content))

		// タイマーを別スレッド（ゴールーチン）で実行
		time.AfterFunc(time.Duration(minutes)*time.Minute, func() {
			reminderMsg := fmt.Sprintf("🔔 <@%s> %d分経ちました！\n設定した内容: **%s**", m.Author.ID, minutes, content)
			s.ChannelMessageSend(m.ChannelID, reminderMsg)
		})
		return
	}
}