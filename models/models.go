package models

import (
	"github.com/watain666/ptt-alertor/models/article"
	"github.com/watain666/ptt-alertor/models/board"
	"github.com/watain666/ptt-alertor/models/user"
)

var User = func() *user.User {
	return user.NewUser(new(user.SQLite))
}
var Article = func() *article.Article {
	return article.NewArticle(new(article.SQLite))
}
var Board = func() *board.Board {
	return board.NewBoard(new(board.SQLite), new(board.SQLite))
}
