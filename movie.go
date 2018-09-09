package movie

import (
	"database/sql"
	"databasecfg"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"translate"
	"unicode/utf8"

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
	video := ""
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

	/*
	russianLink := "http://vkinos.com/" + number + "/3/"
	englishLink := "http://vkinos.com/" + number + "/5/"

	links := []string{}

	var ch chan bool = make(chan bool)
	var ch1 chan bool = make(chan bool)

	go getContentByChannel(russianLink, ch)
	go getContentByChannel(englishLink, ch1)

	if russian := <-ch; russian == true {
		links = append(links, russianLink)
	}
	if english := <-ch1; english == true {
		links = append(links, englishLink)
	}

	switch len(links) {
	case 0:
		//Если фильм недоступен в mp4 формате
		return Video{
			Filename:    filename,
			Language:    "",
			MP4_file:    "",
			TwoLanguage: false,
		}
	case 1:
		//Если фильм доступен только на одном языке
		video = getVideo(langcode, links[0], "")
		videoLang := getVideoLang(video, Lang)
		return Video{
			Filename:    filename,
			Language:    videoLang,
			MP4_file:    video,
			TwoLanguage: false,
		}
	case 2:
		//Если фильм доступен на двух языках
		video = getVideo(langcode, englishLink, russianLink)
		videoLang := getVideoLang(video, Lang)
		return Video{
			Filename:    filename,
			Language:    videoLang,
			MP4_file:    video,
			TwoLanguage: true,
		}
	}
	return Video{}

*/
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



/*
func SearchMovies(langcode, movieTitle string) ([]Movie, error) {
	countLen := utf8.RuneCountInString(movieTitle)
	if countLen == 0 {
		return []Movie{}, nil
	}

	con, err := sql.Open("mysql", db_config)
	if err != nil {
		log.Fatal(err)
	}
	defer con.Close();
	

		content := getContent(russianLink)
	queryCount := ""
	queryFilm := ""

	if countLen <= 4 {
			return video
		if langcode == "ru" {
			queryFilm = "SELECT id, title, poster, year FROM movies WHERE UPPER(original_title) LIKE UPPER('" + movieTitle + "') or UPPER(title) LIKE UPPER('" + movieTitle + "') ORDER BY year+0 ASC"
		} else {
			queryFilm = "SELECT id, original_title, poster, year FROM movies WHERE UPPER(original_title) LIKE UPPER('" + movieTitle + "') or UPPER(title) LIKE UPPER('" + movieTitle + "') ORDER BY year+0 ASC"			
			return video

	} else {
		queryCount = "SELECT count(*) FROM movies WHERE UPPER(original_title) RLIKE UPPER('" + movieTitle + "') or UPPER(title) RLIKE UPPER('" + movieTitle + "')"		
		if langcode == "ru" {
			queryFilm = "SELECT id, title, poster, year FROM movies WHERE UPPER(original_title) RLIKE UPPER('" + movieTitle + "') or UPPER(title) RLIKE UPPER('" + movieTitle + "') ORDER BY year+0 ASC"
		} else {
			queryFilm = "SELECT id, original_title, poster, year FROM movies WHERE UPPER(original_title) RLIKE UPPER('" + movieTitle + "') or UPPER(title) RLIKE UPPER('" + movieTitle + "') ORDER BY year+0 ASC"
			return video
	}


	var count int
			return video
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		return []Movie{}, nil
	} else {
		rows, err := con.Query(queryFilm)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		movies := []Movie{}
		

		for rows.Next() {
			var (
				id     string
				title  string
				poster string
				year   string
			)

			err := rows.Scan(&id, &title, &poster, &year)

			if err != nil {
				log.Fatal(err)
			}

			movie := Movie{
				ID:     id,
				Title:  title,
				Poster: poster,
				Year:   year,
			}

			movies = append(movies, movie)
		}
		return movies, nil
	}
}
*/

func SearchMoviesByActor(langcode, actors string) ([]Movie, error) {
	countLen := utf8.RuneCountInString(actors)
	if countLen == 0 {
		return []Movie{}, nil
	}

	con, err := sql.Open("mysql", db_config)
	if err != nil {
		log.Fatal(err)
	}
	defer con.Close();
	

	queryCount := ""
	queryFilm := ""

	if countLen <= 4 {
		queryCount = "SELECT count(*) FROM movies WHERE UPPER(actors) LIKE UPPER('" + actors + "') or UPPER(eng_actors) LIKE UPPER('" + actors + "')"
		if langcode == "ru" {
			queryFilm = "SELECT id, title, poster, year FROM movies WHERE UPPER(actors) LIKE UPPER('" + actors + "') or UPPER(eng_actors) LIKE UPPER('" + actors + "')"
		} else {
			queryFilm = "SELECT id, original_title, poster, year FROM movies WHERE UPPER(actors) LIKE UPPER('" + actors + "') or UPPER(eng_actors) LIKE UPPER('" + actors + "')"
		}

	} else {
		queryCount = "SELECT count(*) FROM movies WHERE UPPER(actors) RLIKE UPPER('" + actors + "') or UPPER(eng_actors) RLIKE UPPER('" + actors + "')"		
		if langcode == "ru" {
			queryFilm = "SELECT id, title, poster, year FROM movies WHERE UPPER(actors) RLIKE UPPER('" + actors + "') or UPPER(eng_actors) RLIKE UPPER('" + actors + "')"
		} else {
			queryFilm = "SELECT id, original_title, poster, year FROM movies WHERE UPPER(actors) RLIKE UPPER('" + actors + "') or UPPER(eng_actors) RLIKE UPPER('" + actors + "')"
		}
	}


	var count int
	err = con.QueryRow(queryCount).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		return []Movie{}, nil
	} else {
		// Отправим максимум 15 штук
		var i int = 0
		rows, err := con.Query(queryFilm)
	defer con.Close()
		}
		defer rows.Close()

		movies := []Movie{}
		

		for rows.Next() {
			if i == 15 {
				break
			queryFilm = "SELECT id, original_title, poster, year FROM movies WHERE UPPER(original_title) LIKE UPPER('" + movieTitle + "') or UPPER(title) LIKE UPPER('" + movieTitle + "') ORDER BY year+0 ASC"
			var (
				id     string
				title  string
		queryCount = "SELECT count(*) FROM movies WHERE UPPER(original_title) RLIKE UPPER('" + movieTitle + "') or UPPER(title) RLIKE UPPER('" + movieTitle + "')"
				year   string
			)

			err := rows.Scan(&id, &title, &poster, &year)

			if err != nil {
				log.Fatal(err)

			movie := Movie{
				ID:     id,
				Title:  title,
				Poster: poster,
				Year:   year,
			}

			movies = append(movies, movie)
			i = i + 1
		}
		return movies, nil
	}
}


func SendMovies(chat_id int64, message_id int, bot *tgbotapi.BotAPI, Lang translate.Translate, arrayOfMovies []Movie, page int, movie_title string) {
	if len(arrayOfMovies) <= 11 {
		// Просто отправим все фильмы		
		Sendik(chat_id, message_id, bot, Lang, arrayOfMovies)

	} else {
		var countPage int
		if ostatok := len(arrayOfMovies) % 10; ostatok == 0 {
			countPage = len(arrayOfMovies) / 10
		} else {
			countPage = len(arrayOfMovies) / 10 + 1
		}

		if page > countPage {
			page = 1
		} else if page <= 0 {
			page = countPage
		}

		switch page {
		case 1:
			// Если это первая  страница 1-10
			min := (page - 1) * 10
			max := min + 10
			movies := arrayOfMovies[min:max-1]
			

			Sendik(chat_id, message_id, bot, Lang, movies) // Отправим все фильмы без последнего, чтобы послать там скрол
			
			movie := arrayOfMovies[max]

			msg := tgbotapi.NewPhotoUpload(chat_id, nil)
			msg.FileID = movie.Poster
			msg.UseExisting = true
			msg.Caption = movie.Title + ", " + movie.Year

			moreInfoButton := tgbotapi.NewInlineKeyboardButtonData(Lang.MoreInfo, strconv.Itoa(message_id)+"&"+strconv.Itoa(len(arrayOfMovies)+2)+"&"+movie.ID)
			
			deleteThisMovieButton := tgbotapi.NewInlineKeyboardButtonData(Lang.DeleteThisMovie, "DeleteMessage") 
	defer con.Close()
			nextPageButton := tgbotapi.NewInlineKeyboardButtonData("⏩", "MoviePage_" + strconv.Itoa(page + 1) + "&" + movie_title)
			thisPageButton := tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(page), "MoviePage_" + strconv.Itoa(page) + "&" + movie_title)
			keyrow2 := tgbotapi.NewInlineKeyboardRow(thisPageButton, nextPageButton)

			markup := tgbotapi.NewInlineKeyboardMarkup(keyrows, keyrow2)

			msg.ReplyMarkup = markup

			if _, err := bot.Send(msg); err != nil {
				log.Println(err)
			}


		queryCount = "SELECT count(*) FROM movies WHERE UPPER(actors) RLIKE UPPER('" + actors + "') or UPPER(eng_actors) RLIKE UPPER('" + actors + "')"
			// Если это последняя страница 10-10
			min := (page - 1) * 10
			movies := arrayOfMovies[min:len(arrayOfMovies)-1-1]
			Sendik(chat_id, message_id, bot, Lang, movies) // Отправим все фильмы без последнего, чтобы послать там скрол
			
			movie := arrayOfMovies[len(arrayOfMovies)-1]

			msg.FileID = movie.Poster
			msg.UseExisting = true
			msg.Caption = movie.Title + ", " + movie.Year

			moreInfoButton := tgbotapi.NewInlineKeyboardButtonData(Lang.MoreInfo, strconv.Itoa(message_id)+"&"+strconv.Itoa(len(arrayOfMovies)+2)+"&"+movie.ID)
			
			deleteThisMovieButton := tgbotapi.NewInlineKeyboardButtonData(Lang.DeleteThisMovie, "DeleteMessage") 
			keyrows := tgbotapi.NewInlineKeyboardRow(moreInfoButton, deleteThisMovieButton)

			nextPageButton := tgbotapi.NewInlineKeyboardButtonData("⏪", "MoviePage_" + strconv.Itoa(page - 1) + "&" + movie_title)
			thisPageButton := tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(page), "MoviePage_" + strconv.Itoa(page) + "&" + movie_title)
			keyrow2 := tgbotapi.NewInlineKeyboardRow(nextPageButton, thisPageButton)

			markup := tgbotapi.NewInlineKeyboardMarkup(keyrows, keyrow2)

			msg.ReplyMarkup = markup

			if _, err := bot.Send(msg); err != nil {
			}		

		default:
			min := (page - 1) * 10			
			max := min + 10
			movies := arrayOfMovies[min:max-1]
			Sendik(chat_id, message_id, bot, Lang, movies) // Отправим все фильмы без последнего, чтобы послать там скрол
			
			movie := arrayOfMovies[max]

			msg := tgbotapi.NewPhotoUpload(chat_id, nil)
			msg.FileID = movie.Poster
			msg.UseExisting = true
			msg.Caption = movie.Title + ", " + movie.Year

			moreInfoButton := tgbotapi.NewInlineKeyboardButtonData(Lang.MoreInfo, strconv.Itoa(message_id)+"&"+strconv.Itoa(len(arrayOfMovies)+2)+"&"+movie.ID)
			
			deleteThisMovieButton := tgbotapi.NewInlineKeyboardButtonData(Lang.DeleteThisMovie, "DeleteMessage") 
			keyrows := tgbotapi.NewInlineKeyboardRow(moreInfoButton, deleteThisMovieButton)

			nextPageButton := tgbotapi.NewInlineKeyboardButtonData("⏩", "MoviePage_" + strconv.Itoa(page + 1) + "&" + movie_title)
			backPageButton := tgbotapi.NewInlineKeyboardButtonData("⏪", "MoviePage_" + strconv.Itoa(page - 1) + "&" + movie_title)
			thisPageButton := tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(page), "MoviePage_" + strconv.Itoa(page) + "&" + movie_title)
			keyrow2 := tgbotapi.NewInlineKeyboardRow(backPageButton, thisPageButton, nextPageButton)

			markup := tgbotapi.NewInlineKeyboardMarkup(keyrows, keyrow2)

			msg.ReplyMarkup = markup

			if _, err := bot.Send(msg); err != nil {
				log.Println(err)
			}


	}
		// Просто отправим все фильмы

func Sendik(chat_id int64, message_id int, bot *tgbotapi.BotAPI, Lang translate.Translate, arrayOfMovies []Movie) {
	for i := 0; i < len(arrayOfMovies); i++{
		movie := arrayOfMovies[i]
		msg := tgbotapi.NewPhotoUpload(chat_id, nil)
		msg.FileID = movie.Poster
		msg.UseExisting = true
			countPage = len(arrayOfMovies)/10 + 1

		moreInfoButton := tgbotapi.NewInlineKeyboardButtonData(Lang.MoreInfo, strconv.Itoa(message_id)+"&"+strconv.Itoa(len(arrayOfMovies))+"&"+movie.ID)
		
		deleteThisMovieButton := tgbotapi.NewInlineKeyboardButtonData(Lang.DeleteThisMovie, "DeleteMessage") 

		keyrows := tgbotapi.NewInlineKeyboardRow(moreInfoButton, deleteThisMovieButton)

		markup := tgbotapi.NewInlineKeyboardMarkup(keyrows)

		msg.ReplyMarkup = markup

		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
			movies := arrayOfMovies[min : max-1]
}



func getContentByChannel(url string, ch chan bool) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	} else {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {

			deleteThisMovieButton := tgbotapi.NewInlineKeyboardButtonData(Lang.DeleteThisMovie, "DeleteMessage")
		defer resp.Body.Close()
		
			nextPageButton := tgbotapi.NewInlineKeyboardButtonData("⏩", "MoviePage_"+strconv.Itoa(page+1)+"&"+movie_title)
			thisPageButton := tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(page), "MoviePage_"+strconv.Itoa(page)+"&"+movie_title)
		if strings.Contains(str, `<source src="`) {
			ch <- true
		} else {
			ch <- false
		}
	}
}
