# spodlivoibot

Bot API Key:
First of all, create bot with [@BotFather](https://t.me/BotFather).  
Next you have two way: 
1. Fill your values in main.go
2. Fill your values in command.
Command example:
`BOT_KEY='your key' go run github.com/Aryesia/spodlivoi_bot`
`BOT_KEY='your key' ./spodlivoi_go_bot`

Build: 
1. `go get github.com/go-telegram-bot-api/telegram-bot-api`
2. `go get github.com/mattn/go-sqlite3`
3. `go get github.com/tidwall/gjson`
4. `go get github.com/Aryesia/spodlivoi_bot`
5. `go build github.com/Aryesia/spodlivoi_bot`  
6. `go install github.com/Aryesia/spodlivoi_bot`  
7. mv go/src/github.com/Aryesia/spodlivoi_go_bot to directory with DB and .txt files
8. `./.../spodlivoi_go_bot` 

Run:  
`go run github.com/Aryesia/spodlivoi_bot`

Demo: [@spodlivoi_bot](https://t.me/spodlivoi_bot)  
