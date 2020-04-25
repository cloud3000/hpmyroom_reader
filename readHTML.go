package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"golang.org/x/net/html"
)

type post struct {
	myRoom    string
	sigString string
	tStamp    string
	content   []string
}

func main() {
	os.Remove("./foo.db")

	var name string = "ESDomain"
	r := getRoombody(name)
	cnt, err := insertBodyPosts(name, r)
	if err != nil {
		fmt.Println("Error occurred: ", err)
	} else {
		fmt.Println(cnt, "Posts Inserted from", name)
	}
	return
}
func getRoombody(roomname string) *http.Response {
	resp, err := http.Get("http://localhost:8080/ESDomain.html")
	if err != nil {
		fmt.Println("Cannot read HTML source code.")
		log.Fatal(err)
	}
	return resp
}

func insertBodyPosts(name string, r *http.Response) (int, error) {
	z := html.NewTokenizer(r.Body)
	fmt.Println(z)
	var newPost bool = false
	var header bool = false
	var newCount, contIdx, postcnt int
	var myPost post
	var err error
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	sqlStmt := `
	CREATE TABLE [IF NOT EXISTS] myroom_posts (
		id integer PRIMARY KEY,
		room_name text NOT NULL,
		author text NOT NULL,
		time_stamp text NOT NULL,
		content text NOT NULL,
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	for {

		//fmt.Println("Inside for loop")

		tt := z.Next()
		toeKen := z.Token()
		//fmt.Println(" ===============> ", tt)
		if newPost {
			header = true
			if len(myPost.content) > 0 {
				myPost.myRoom = name
				postcnt++
				fmt.Println(myPost)
			}
			//fmt.Println(newCount%2, toeKen.Data)
			if newCount%2 == 0 {
				myPost.sigString = toeKen.Data
			} else {
				myPost.tStamp = toeKen.Data
			}
			newPost = false
			newCount++
			myPost.content = nil
			contIdx = 0
		} else {
			header = false
		}

		//fmt.Println(toeKen.Data)
		if tt == html.ErrorToken {
			// End of the document, we're done
			return postcnt, err
		}
		if tt == html.StartTagToken {
			//fmt.Println(toeKen.Data)
			// for _, v := range toeKen.Attr {
			// 	fmt.Printf("%s=[%s]\n", toeKen.Data, v)
			// }
			if toeKen.Data == "span" && len(toeKen.Attr) > 0 {
				newPost = fmt.Sprintf("%s", toeKen.Attr[0]) == "{ style  color:#773c00;}" // [{ style  color:#773c00;}]
			}
		}
		if tt == html.TextToken && header == false {
			str := fmt.Sprintf("%s", toeKen.Data)
			if len(strings.TrimSpace(str)) > 0 {
				myPost.content = append(myPost.content, fmt.Sprintf("%s", strings.TrimSpace(str)))
				contIdx++
			}
		}
	} // End for loop
	return postcnt, err
}
