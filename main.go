package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
)

var botAPIKey = "" //paste your bot token here

var brepairText = "ЧИНИ "
var repairText = strings.Repeat(brepairText, 1000)

var stickerPacks = []string{"sosatlezhatsosat", "fightpics", "test228idinaxui", "davlyu", "gasiki2", "durkaebt", "daEntoOn", "Bodyafleks3"}

var path, pathError = os.Getwd()

func main() {
	if pathError != nil {
		log.Panic(pathError)
	}
	// if strings.Contains(path, "spodlivoi_go_bot") {
	// 	path = strings.ReplaceAll(path, "/spodlivoi_go_bot", "")
	// }
	if botAPIKey == "" {
		if os.Getenv("BOT_KEY") != "" {
			botAPIKey = os.Getenv("BOT_KEY")
		} else if os.Args[1] != "" {
			botAPIKey = os.Args[1]
		} else {
			log.Printf("API key missing!")
			return
		}
	}

	bot, err := tgbotapi.NewBotAPI(botAPIKey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)
	log.Printf("Current working directory %s", path)

	db, err := sql.Open("sqlite3", path+"/db/spodlivoi.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbCreate(db)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	log.Printf("Bot is running")
	for update := range updates {
		if update.Message != nil && update.Message.IsCommand() {
			names := strings.Split(update.Message.Text, "@")
			if len(names) > 1 && names[1] != bot.Self.UserName {
				continue
			}
			register(db, update, bot)
			switch update.Message.Command() {
			case "test":
				sendMessageWithReply(update, bot, "Я работаю, а твой писюн - нет!")
				break
			case "dick":
				rollDick(db, update, bot)
				break
			case "fight":
				sendFightSticker(update, bot, false)
				break
			case "baby", "dota", "olds", "kolchan", "shizik":
				sendRandomCopypaste(update.Message.Command(), update, bot)
				break
			case "voice":
				sendVoice(update, bot)
				break
			case "add_voice":
				addVoice(update, bot)
				break
			case "del_voice":
				delVoice(update, bot)
				break
			case "bred":
				go getBred(update, bot)
				break
			case "webm":
				go sendRandomWebm(update, bot)
				break
			case "topdicks":
				sendTopDicks(db, update, bot)
				break
			default:
				sendFightSticker(update, bot, true)
				break
			}
		} else if update.Message != nil {
			if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.UserName == bot.Self.UserName {
				sendFightSticker(update, bot, true)
				break
			}
		} else if update.EditedMessage != nil {
			msg := tgbotapi.NewMessage(update.EditedMessage.Chat.ID, "Анус себе отредактируй!")
			msg.ReplyToMessageID = update.EditedMessage.MessageID
			bot.Send(msg)
		} else if update.InlineQuery != nil {
			log.Printf(update.InlineQuery.Query)
			article1 := tgbotapi.NewInlineQueryResultArticleMarkdown("1", "Олды", getRandomCopypaste("olds"))
			article1.Description = "Платиновые пасты дотатреда"
			article2 := tgbotapi.NewInlineQueryResultArticleMarkdown("2", "Дота", getRandomCopypaste("dota"))
			article2.Description = "Малолетние дэбилы"
			article3 := tgbotapi.NewInlineQueryResultArticleMarkdown("3", "Ребёнок", getRandomCopypaste("baby"))
			article3.Description = "Ещё малолетние дэбилы"
			article4 := tgbotapi.NewInlineQueryResultArticleMarkdown("4", "Колчан", getRandomCopypaste("kolchan"))
			article4.Description = "Зачем мой хуй перешёл тебе в рот?"
			article5 := tgbotapi.NewInlineQueryResultArticleMarkdown("5", "Шизик", getRandomCopypaste("shizik"))
			article5.Description = "T9 insanity"
			articleRepair := tgbotapi.NewInlineQueryResultArticleMarkdown("6", "ЧИНИ", repairText)
			articleRepair.Description = "ЧИНИ ЧИНИ ЧИНИ"
			var results []interface{}
			results = append(results, article1, article2, article3, article4, article5, articleRepair)
			inlineConf := tgbotapi.InlineConfig{
				InlineQueryID: update.InlineQuery.ID,
				IsPersonal:    true,
				CacheTime:     0,
				Results:       results,
			}
			bot.AnswerInlineQuery(inlineConf)
		}

	}
}

func sendTopDicks(db *sql.DB, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	message := "Топ писюнов:\n\n"
	var chatID int64
	db.QueryRow("select id from chats where chat_id = ?", update.Message.Chat.ID).Scan(&chatID)
	rows, err := db.Query("select id, user_name from users where chat = ?", chatID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	dicks := make([]Dicks, 0)
	for rows.Next() {
		var userID int
		var size int
		var userName string
		rows.Scan(&userID, &userName)
		db.QueryRow("select size from dicks where user = ?", userID).Scan(&size)
		dicks = append(dicks, Dicks{userName: userName, size: size})
	}
	sort.Slice(dicks, func(i, j int) bool {
		return dicks[i].size < dicks[j].size
	})
	for i, d := range dicks {
		message += fmt.Sprintf("%d. "+d.userName+" - %dсм;\n", i+1, d.size)
	}
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, message))
}

//Dicks model
type Dicks struct {
	userName string
	size     int
}

func sendVoice(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	dat, _ := ioutil.ReadFile(path + "/db/voice")
	datS := strings.ReplaceAll(string(dat), "\n", "")
	if datS == "" {
		sendMessageWithReply(update, bot, "Ты еблан? Голосовые добавь!")
		return
	}
	data := strings.Split(datS, ";")
	if len(data) == 0 {
		sendMessageWithReply(update, bot, "Ты еблан? Голосовые добавь!")
		return
	}
	number := getRandomNumberInRange(0, len(data)-1)
	if data[1] == "" {
		number = 0
	}
	voice := tgbotapi.NewVoiceShare(update.Message.Chat.ID, data[number])
	if update.Message.ReplyToMessage != nil {
		voice.ReplyToMessageID = update.Message.ReplyToMessage.MessageID
		bot.Send(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
	}
	bot.Send(voice)
}

func addVoice(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.Voice != nil {
		voice := update.Message.ReplyToMessage.Voice
		dat, _ := ioutil.ReadFile("db/voice")
		datS := strings.ReplaceAll(string(dat), "\n", "")
		data := strings.Split(datS, ";")
		item := stringInSlice(voice.FileID, data)
		if item != -1 {
			sendMessageWithReply(update, bot, "Ты еблан? Нахуя мне повтор нужен?")
			return
		}
		f, _ := os.OpenFile("db/voice", os.O_APPEND|os.O_WRONLY, 0644)
		defer f.Close()
		f.WriteString(voice.FileID + ";")
		sendMessageWithReply(update, bot, "Добавлено!")
	} else {
		sendMessageWithReply(update, bot, "Ты еблан? А голосовое то где?")
	}
}

func delVoice(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.Voice != nil {
		voice := update.Message.ReplyToMessage.Voice
		dat, _ := ioutil.ReadFile("db/voice")
		datS := strings.ReplaceAll(string(dat), "\n", "")
		if datS == "" {
			sendMessageWithReply(update, bot, "Ты еблан? Голосовые добавь!")
			return
		}
		data := strings.Split(datS, ";")
		if len(data) == 0 {
			sendMessageWithReply(update, bot, "Ты еблан? Голосовые добавь!")
			return
		}
		item := stringInSlice(voice.FileID, data)
		if item == -1 {
			sendMessageWithReply(update, bot, "Ты еблан? Нет такого голосового!")
			return
		}
		datS = strings.ReplaceAll(datS, voice.FileID+";", "")
		f, _ := os.Create("db/voice")
		w := bufio.NewWriter(f)
		defer f.Close()
		w.WriteString(datS)
		w.Flush()
		sendMessageWithReply(update, bot, "Удалено!")
	} else {
		sendMessageWithReply(update, bot, "Ты еблан? А голосовое то где?")
	}
}

var dvachURL = "https://2ch.hk/b/catalog.json"

func getBThreads() ([]gjson.Result, error) {
	var client http.Client
	resp, err := client.Get(dvachURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return gjson.Get(string(bodyBytes), "threads").Array(), nil
	}
	return nil, errors.New(string(resp.StatusCode))
}

func getBred(update tgbotapi.Update, bot *tgbotapi.BotAPI) {

	posts, _ := getBThreads()
	number := getRandomNumberInRange(0, len(posts)-1)
	post := gjson.Get(posts[number].String(), "comment").String()
	post = strings.ReplaceAll(post, "<br>", "\n")
	post = strings.ReplaceAll(post, "<b>", "*")
	post = strings.ReplaceAll(post, "</b>", "*")
	post = strings.ReplaceAll(post, "<i>", "_")
	post = strings.ReplaceAll(post, "</i>", "_")
	post = strings.ReplaceAll(post, "<strong>", "*")
	post = strings.ReplaceAll(post, "</strong>", "*")
	post = strings.ReplaceAll(post, "<[a-zA-Z0-9=\\-\".,/_ ]+>", "")
	post += "\n\nЧитать подробнее: https://2ch.hk/b/res/" + gjson.Get(posts[number].String(), "num").String() + ".html"
	post = "_" + gjson.Get(posts[number].String(), "date").String() + "_\n\n" + post
	files := gjson.Get(posts[number].String(), "files").Array()
	if len(files) == 0 || len(post) > 1024 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, post)
		msg.ParseMode = "Markdown"
		msg.DisableWebPagePreview = false
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf(err.Error())
		}
	} else {
		msg := tgbotapi.NewPhotoShare(update.Message.Chat.ID, "https://2ch.hk"+gjson.Get(files[0].String(), "path").String())
		msg.ParseMode = "Markdown"
		msg.Caption = post
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf(err.Error())
		}

	}

}

func getWebmURL() (string, error) {
	posts, _ := getBThreads()
	threadNumber := ""
	for _, res := range posts {
		subject := strings.ToLower(gjson.Get(res.String(), "subject").String())
		if strings.Contains(subject, "webm thread") || strings.Contains(subject, "webm-thread") ||
			strings.Contains(subject, "цуиь thread") || strings.Contains(subject, "цуиь-thread") ||
			strings.Contains(subject, "webm тред") || strings.Contains(subject, "webm-тред") ||
			strings.Contains(subject, "цуиь тред") || strings.Contains(subject, "цуиь-тред") {
			threadNumber = gjson.Get(res.String(), "num").String()
			break
		}
	}
	if threadNumber == "" {
		return "", nil
	}
	var client http.Client
	resp, err := client.Get("https://2ch.hk/makaba/mobile.fcgi?task=get_thread&board=b&thread=" + threadNumber + "&post=0")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "nil", err
		}
		log.Printf("https://2ch.hk/makaba/mobile.fcgi?task=get_thread&board=b&thread=" + threadNumber + "&post=0")
		posts := gjson.Parse(string(bodyBytes)).Array()
		file := ""
		for file == "" {
			number := getRandomNumberInRange(0, len(posts)-1)
			files := gjson.Get(posts[number].String(), "files").Array()
			if len(files) != 0 {
				file = "https://2ch.hk" + gjson.Get(files[0].String(), "path").String()
			}
			if !strings.Contains(file, ".mp4") && !strings.Contains(file, ".webm") {
				file = ""
			}
		}
		return file, nil
	}
	return "", errors.New(string(resp.StatusCode))
}

func dbCreate(db *sql.DB) {
	dat, _ := ioutil.ReadFile(path + "/db/podliva.sql")
	_, err := db.Exec(string(dat))
	if err != nil {
		log.Printf("%v", err)
	}
}

func sendRandomWebm(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	video, err := getWebmURL()
	if err != nil || video == "" {
		sendMessageWithReply(update, bot, "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
		return
	}
	if strings.Contains(video, ".webm") {
		resp, _ := http.Get(video)
		filename := string(getRandomNumberInRange(0, 1000000000))
		defer resp.Body.Close()
		out, _ := os.Create("/tmp/" + filename + ".webm")
		defer out.Close()
		io.Copy(out, resp.Body)
		err := exec.Command("ffmpeg", "-hide_banner -i /tmp/"+filename+".webm "+"-acodec copy -vcodec copy -strict -2 -f mp4 /tmp/"+filename+".mp4")
		if err != nil {
			log.Printf(err.String())
			sendMessageWithReply(update, bot, "Произошёл пиздос")
			return
		}
		video = "/tmp/" + filename + ".mp4"
		mes := tgbotapi.NewVideoUpload(update.Message.Chat.ID, video)
		mes.ReplyToMessageID = update.Message.MessageID
		_, errr := bot.Send(mes)
		if errr != nil {
			sendMessageWithReply(update, bot, "Какой-то уебан закодировал видео в vp8...\nПодожди. Мне нужно немного времени, чтоб исправить это.")
			exec.Command("rm -rf /tmp/" + filename + ".mp4")
			exec.Command(
				"ffmpeg -hide_banner -i /tmp/" + filename + ".webm " +
					"-acodec copy -vcodec libx264 -strict -2 -f mp4 /tmp/" + filename + ".mp4")
			if err != nil {
				log.Printf(err.String())
				sendMessageWithReply(update, bot, "Произошёл пиздос")
				return
			}
			mes := tgbotapi.NewVideoUpload(update.Message.Chat.ID, video)
			mes.ReplyToMessageID = update.Message.MessageID
			bot.Send(mes)
		}
	} else {
		mes := tgbotapi.NewVideoShare(update.Message.Chat.ID, video)
		mes.ReplyToMessageID = update.Message.MessageID
		bot.Send(mes)
	}
}

func removeFromStringArray(s []string, i int) []string {

	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func stringInSlice(a string, list []string) int {
	for i, b := range list {
		if b == a {
			return i
		}
	}
	return -1
}

func getRandomCopypaste(name string) string {
	dat, _ := ioutil.ReadFile(path + "/res/" + name + ".txt")
	data := strings.Split(string(dat), "|")
	number := getRandomNumberInRange(0, len(data)-1)
	return data[number]
}

func sendRandomCopypaste(name string, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, getRandomCopypaste(name))
	if update.Message.ReplyToMessage != nil {
		msg.ReplyToMessageID = update.Message.ReplyToMessage.MessageID
		bot.Send(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
	}
	bot.Send(msg)
}

func errorMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI, err error) {
	sendMessageWithReply(update, bot, "Ебучий разработчик опять допустил ошибку и нихуя не работает")
	log.Printf(err.Error())
}

func sendMessageWithReply(update tgbotapi.Update, bot *tgbotapi.BotAPI, text string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyToMessageID = update.Message.MessageID
	bot.Send(msg)
}

func rollDick(db *sql.DB, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var chatID int64
	db.QueryRow("select id from chats where chat_id = ?", update.Message.Chat.ID).Scan(&chatID)
	var userID int64
	db.QueryRow("select id from users where user_id = ? AND chat = ?", update.Message.From.ID, chatID).Scan(&userID)
	var count int64
	size := 0
	db.QueryRow("select COUNT(*) from dicks where user = ?", userID).Scan(&count)
	first := count == 0
	if !first {
		var last time.Time
		now := time.Now()
		db.QueryRow("select last_measurement from dicks where user = ?", userID).Scan(&last)
		if last.Day() == now.Day() && last.Month() == now.Month() {
			sendMessageWithReply(update, bot, fmt.Sprintf("Ты уже измерял свой огрызок сегодня!\nПриходи через %dч %dм", 23-last.Hour(), 59-last.Minute()))
			return
		}
		db.QueryRow("select size from dicks where user = ?", userID).Scan(&size)
	}

	upSize := getRandomNumberInRange(-15, 15)

	if first {
		if upSize > 0 {
			size += upSize
		}
		_, err := db.Exec("insert into dicks(user, size, last_measurement)VALUES(?, ?, ?)", userID, size, time.Now())
		if err != nil {
			log.Fatal(err.Error())
		}
		if size == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "На данный момент ты не имеешь писюна, неудачник!")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ого! Размер твоего писюна аж %dсм!\nПриходи завтра. Посмотрим, изменился ли он", size))
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	} else {
		if upSize == 0 || size+upSize <= 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Мои соболезнования. Сегодня у тебя произошла страшная трагедия: твой писюн отпал.")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			size = 0
		} else if upSize > 0 {
			size += upSize
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Поздравляю! Твой член сегодня вырос на целых %dсм.\nТеперь его длина %dсм. Скоро ты станешь настоящим мужчиной!", upSize, size))
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)

		} else {
			size += upSize
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ахахахах, неудачник. Твой огрызок стал меньше на целых %dсм.\nТеперь его длина %dсм.", upSize, size))
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}

		db.Exec("update dicks set size = ? , last_measurement = ?", size, time.Now())
	}
}

func sendFightSticker(update tgbotapi.Update, bot *tgbotapi.BotAPI, reply bool) {
	number := getRandomNumberInRange(0, len(stickerPacks)-1)
	set, _ := bot.GetStickerSet(tgbotapi.GetStickerSetConfig{Name: stickerPacks[number]})
	number = getRandomNumberInRange(0, len(set.Stickers)-1)
	sticker := set.Stickers[number]
	msg := tgbotapi.NewStickerShare(update.Message.Chat.ID, sticker.FileID)
	if reply {
		msg.ReplyToMessageID = update.Message.MessageID
	} else if update.Message.ReplyToMessage != nil {
		msg.ReplyToMessageID = update.Message.ReplyToMessage.MessageID
		bot.Send(tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
	}
	bot.Send(msg)
}

func register(db *sql.DB, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var count int64
	db.QueryRow("select COUNT(*) from chats where chat_id = ?", update.Message.Chat.ID).Scan(&count)
	if count == 0 {
		db.Exec("insert into chats(chat_id, chat_name)VALUES(?, ?)", update.Message.Chat.ID, update.Message.Chat.Title)
	}
	var chatID int64
	db.QueryRow("select id from chats where chat_id = ?", update.Message.Chat.ID).Scan(&chatID)
	var userCount int64
	db.QueryRow("select COUNT(*) from users where user_id = ? AND chat = ?", update.Message.From.ID, chatID).Scan(&userCount)
	if userCount == 0 {
		db.Exec("insert into users(user_id, user_name, chat)VALUES(?, ?, ?)",
			update.Message.From.ID, update.Message.From.UserName, chatID)
	}
}

func getRandomNumberInRange(min, max int) (number int) {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}
