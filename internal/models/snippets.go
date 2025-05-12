package models

import (
	"database/sql"
	"errors"
	"time"
)

type SnippetModelInterface interface {
	Insert(title string, content string, expires int) (int, error)
	Get(id int) (Snippet, error)
	Latest() ([]Snippet, error)
}

type Snippet struct {
	Id      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert(title, content string, expires int) (int, error) {
	statement := "insert into snippets(title, content, created, expires) values ($1, $2, now(), now() + make_interval(days => $3)) returning id"
	var id int
	err := m.DB.QueryRow(statement, title, content, expires).Scan(&id)

	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *SnippetModel) Get(id int) (Snippet, error) {
	statement := "select id, title, content, created, expires from snippets where expires > now() and id = $1"
	row := m.DB.QueryRow(statement, id)
	var snippet Snippet
	err := row.Scan(&snippet.Id, &snippet.Title, &snippet.Content, &snippet.Created, &snippet.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		}
		return Snippet{}, err
	}

	return snippet, nil
}

func (m *SnippetModel) Latest() ([]Snippet, error) {
	statement := "select id, title, content, created, expires from snippets where expires > now() order by id desc limit 10"
	rows, err := m.DB.Query(statement)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	snippets := make([]Snippet, 0, 10)
	for rows.Next() {
		var snippet Snippet

		err = rows.Scan(&snippet.Id, &snippet.Title, &snippet.Content, &snippet.Created, &snippet.Expires)

		if err != nil {
			return nil, err
		}

		snippets = append(snippets, snippet)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
