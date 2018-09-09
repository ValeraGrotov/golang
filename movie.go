package movie

import (
	"database/sql"
	"databasecfg"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"translate"
	
	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/go-sql-driver/mysql"
)

var (
	db_config = databasecfg.GetConfig()
)

type Movie struct {
	ID        string
	Type      string
	Poster    string
	Title     string
	Genre     string
	Year      string
	Actors    string
	Videofile string
	Filename  string
}

type Video struct {
	Filename    string
	Language    string
	MP4_file    string
	TwoLanguage bool
	IframeLink  string
}

func (movie Movie) SendMovie(chat_id int64, bot *tgbotapi.BotAPI, langcode string, Lang translate.Translate) {

	if movie.Type == "SERIES" {
		// Если это сериал.
		video := movie.Videofile
		text := "<b>Год</b>\n" + movie.Year + "\n\n<b>Жанр</b>\n" + movie.Genre + "\n\n<b>Актеры</b>\n" + movie.Actors

		keyboard := tgbotapi.NewInlineKeyboardButtonURL("Смотреть в браузере", video)
		showMenuButton := tgbotapi.NewInlineKeyboardButtonData(Lang.ShowMenuButton, "showMenuButton")

		keyrows := tgbotapi.NewInlineKeyboardRow(keyboard)
		closeMovieButtonRow := tgbotapi.NewInlineKeyboardRow(showMenuButton)
		markup := tgbotapi.NewInlineKeyboardMarkup(keyrows, closeMovieButtonRow)

		photo := tgbotapi.NewPhotoUpload(chat_id, nil)
		photo.FileID = movie.Poster
		photo.UseExisting = true
		photo.Caption = movie.Title

		msg := tgbotapi.NewMessage(chat_id, text)
		msg.ParseMode = "html"
		msg.ReplyMarkup = markup

		if _, err := bot.Send(photo); err != nil {
			log.Println(err)
		}
		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}
	} else {
		video := NewVideo(langcode, movie.Filename, Lang)

		if video.MP4_file == "" {
			// Если фильм не удалось найти
			photo := tgbotapi.NewPhotoUpload(chat_id, nil)
			photo.FileID = movie.Poster
			photo.UseExisting = true
			photo.Caption = movie.Title + ", " + movie.Year
			msg := tgbotapi.NewMessage(chat_id, "<b>"+Lang.Genre+"</b>\n"+movie.Genre+"\n\n<b>"+Lang.Actors+"</b>\n"+movie.Actors)
			msg.ParseMode = "html"

			keyboard := tgbotapi.NewInlineKeyboardButtonURL("Смотреть на русском", video.IframeLink)
			keyrows := tgbotapi.NewInlineKeyboardRow(keyboard)

			showMenuButton := tgbotapi.NewInlineKeyboardButtonData(Lang.ShowMenuButton, "showMenuButton")

			closeMovieButtonRow := tgbotapi.NewInlineKeyboardRow(showMenuButton)

			trailerButton := tgbotapi.NewInlineKeyboardButtonData(Lang.Trailer, "ShowTrailer"+movie.ID)
			trailerButtonRow := tgbotapi.NewInlineKeyboardRow(trailerButton)

			markup := tgbotapi.NewInlineKeyboardMarkup(keyrows, trailerButtonRow, closeMovieButtonRow)
			msg.ReplyMarkup = markup
			if _, err := bot.Send(photo); err != nil {
				log.Println(err)
			}
			if _, err := bot.Send(msg); err != nil {
				log.Println(err)
			}
		} else {
			photo := tgbotapi.NewPhotoUpload(chat_id, nil)
			photo.FileID = movie.Poster
			photo.UseExisting = true
			photo.Caption = movie.Title + ", " + movie.Year
			msg := tgbotapi.NewMessage(chat_id, "<b>"+Lang.Genre+"</b>\n"+movie.Genre+"\n\n<b>"+Lang.Actors+"</b>\n"+movie.Actors)

			keyboard := tgbotapi.NewInlineKeyboardButtonURL(video.Language, video.MP4_file)
			trailerButton := tgbotapi.NewInlineKeyboardButtonData(Lang.Trailer, "ShowTrailer"+movie.ID)
			trailerButtonRow := tgbotapi.NewInlineKeyboardRow(trailerButton)
			rateButton := tgbotapi.NewInlineKeyboardButtonURL(Lang.MenuRate, "https://t.me/tchannelsbot?start=KinonetDB")
			rateButtonRow := tgbotapi.NewInlineKeyboardRow(rateButton)

			keyrows := tgbotapi.NewInlineKeyboardRow(keyboard)

			showMenuButton := tgbotapi.NewInlineKeyboardButtonData(Lang.ShowMenuButton, "showMenuButton")

			closeMovieButtonRow := tgbotapi.NewInlineKeyboardRow(showMenuButton)

			markup := tgbotapi.NewInlineKeyboardMarkup(keyrows, trailerButtonRow, closeMovieButtonRow)

			if video.TwoLanguage == true {
				markup = tgbotapi.NewInlineKeyboardMarkup(keyrows, trailerButtonRow, closeMovieButtonRow, rateButtonRow)
			}

			msg.ParseMode = "html"
			msg.ReplyMarkup = markup

			if _, err := bot.Send(photo); err != nil {
				log.Println(err)
			}
			if _, err := bot.Send(msg); err != nil {
				log.Println(err)
			}
		}
	}
}

func GetMovieByID(langcode, idFilm string) (Movie, error) {
	con, err := sql.Open("mysql", db_config)
	if err != nil {
		log.Fatal(err)
	}
	defer con.Close()

	movie := Movie{}

	switch langcode {
	case "ru":
		err = con.QueryRow("SELECT id, type, poster, title, year, genre, actors, filename, videofile FROM movies WHERE id=?", idFilm).Scan(&movie.ID, &movie.Type, &movie.Poster, &movie.Title, &movie.Year, &movie.Genre, &movie.Actors, &movie.Filename, &movie.Videofile)
	default:
		err = con.QueryRow("SELECT id, type, poster, original_title, year, eng_genre, eng_actors, filename, videofile FROM movies WHERE id=?", idFilm).Scan(&movie.ID, &movie.Type, &movie.Poster, &movie.Title, &movie.Year, &movie.Genre, &movie.Actors, &movie.Filename, &movie.Videofile)
	}

	if err == nil {
		return movie, nil
	} else if err == sql.ErrNoRows {
		// Нет в базе. Сказать об этом пользователю
		return Movie{}, err

	} else {
		// Другая ошибка
		return Movie{}, err
	}
}

func NewVideo(langcode, filename string, Lang translate.Translate) Video {
	ar := strings.Split(filename, "/")
	ar = strings.Split(ar[2], "-")
	number := ar[0]

	link := "http://vkinos.com/" + number + "/1/"

	movieIframeLink := getMovieIframeLink(link)

	return Video{
		Filename:    filename,
		Language:    "",
		MP4_file:    "",
		TwoLanguage: false,
		IframeLink: movieIframeLink,
	}
}


func getMovieIframeLink(url string) string {
	content := getContent(url)

	if strings.Contains(content, `<iframe src="`) == true {
		ar := strings.Split(content, `<iframe src="`)
		ar = strings.Split(ar[1], "\"")

		movieLink := ar[0] + "?ref=mp4"
		return movieLink
	}

	return "http://vk.com"
}

func getContent(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return ""
	} else {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		resp.Body.Close()
		return string(bytes)
	}
}
