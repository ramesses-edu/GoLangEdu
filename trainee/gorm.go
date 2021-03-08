package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	netResource string = "https://jsonplaceholder.typicode.com/"
	userDB      string = "utest"
	passDB      string = "12345"
	hostDB      string = "localhost"
	portDB      string = "3306"
	nameDB      string = "edudb"
	mux         sync.Mutex
)

type post struct {
	UserID int    `json:"userId" gorm:"column:userId"`
	ID     int    `json:"id" gorm:"column:id;primaryKey"`
	Title  string `json:"title" gorm:"column:title;type:VARCHAR(256)"`
	Body   string `json:"body" gorm:"column:body;type:VARCHAR(256)"`
}
type comment struct {
	PostID int    `json:"postId" gorm:"column:postId"`
	ID     int    `json:"id" gorm:"column:id;primaryKey"`
	Name   string `json:"name" gorm:"column:name;type:VARCHAR(256)"`
	Email  string `json:"email" gorm:"column:email;type:VARCHAR(256)"`
	Body   string `json:"body" gorm:"column:body;type:VARCHAR(256)"`
}

func main1() {
	postsURI := "posts?userId=7"
	url := netResource + postsURI
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	var posts []post
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(respBody, &posts)
	if err != nil {
		fmt.Println(err)
	}
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", userDB, passDB, hostDB, portDB, nameDB)
	gormD := mysql.New(mysql.Config{
		DSN:               dsn,
		DefaultStringSize: 256,
	})
	db, err := gorm.Open(gormD, &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		return
	}
	sql, _ := db.DB()
	defer sql.Close()
	sql.Ping()
	db.Migrator().DropTable(&post{})
	db.Migrator().DropTable(&comment{})
	db.Migrator().CreateTable(&post{})
	db.Migrator().CreateTable(&comment{})
	for _, value := range posts {
		wg.Add(1)
		go func(ctx context.Context, p post, db *gorm.DB) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				mux.Lock()
				db.Create(&p)
				mux.Unlock()
				cmntURI := "comments?postId="
				url = netResource + cmntURI + strconv.Itoa(p.ID)
				fmt.Printf("id: %d; userid: %d \n", p.ID, p.UserID)
				var comments []comment
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer resp.Body.Close()
				respBody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = json.Unmarshal(respBody, &comments)
				if err != nil {
					fmt.Println(err)
				}
				var wg2 sync.WaitGroup
				ctx2, cancel := context.WithTimeout(ctx, time.Second*5)
				defer cancel()
				for _, value := range comments {
					wg2.Add(1)
					go func(ctx context.Context, c comment, db *gorm.DB) {
						defer wg2.Done()
						select {
						case <-ctx.Done():
							return
						default:
							mux.Lock()
							db.Create(&c)
							mux.Unlock()
							fmt.Printf("comId: %d postId: %d \n", c.ID, c.PostID)
						}
					}(ctx2, value, db)
				}
				wg2.Wait()
			}

		}(ctx, value, db)
	}
	wg.Wait()

}
