package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	// mysql connector
	_ "github.com/go-sql-driver/mysql"
	sqlx "github.com/jmoiron/sqlx"

	// for better printing
	"github.com/modood/table"
)

var (
	User     string
	Password string
	DBName   string
)

type Library struct {
	db *sqlx.DB
}

type Books struct {
	Title      string
	ISBN       string
	Author     string
	Publisher  string
	Stock      int
	Available  int
	RemoveInfo sql.NullString
}

type Users struct {
	ID       string
	Name     string
	Password string
	Overdue  int
	Type     int
}

type Records struct {
	recordID    string
	bookID      string
	userID      string
	IsReturned  bool
	borrowDate  time.Time
	returnDate  sql.NullTime
	deadline    time.Time
	extendTimes int
}

var help string
var timeTemplate = "2006/01/02 15:04:05"
var ErrAllRemoved = errors.New("All have been removed.")
var ErrUserExists = errors.New("User account already exists.")
var ErrBookNotExists = errors.New("Book not exists.")
var ErrUserNotExists = errors.New("User not exists.")
var ErrUserSuspended = errors.New("Account suspended.")
var ErrBookNotAvailable = errors.New("There is no available book.")
var ErrAlreadyBorrowed = errors.New("The book is already borrowed.")
var ErrNotBorrowed = errors.New("The book isn't borrowed.")
var ErrNoMoreExtended = errors.New("Already extended for three times. Can't extend again.")
var ErrPassword = errors.New("Username and password don't match")

// ConnectDB make connection to local database
func (lib *Library) ConnectDB() {
	file, err := os.Open("config.ini")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	User = scanner.Text()
	scanner.Scan()
	Password = scanner.Text()
	scanner.Scan()
	DBName = scanner.Text()
	db, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", User, Password, DBName+"?charset=utf8&loc=Asia%2FShanghai&parseTime=true"))
	if err != nil {
		panic(err)
	}
	lib.db = db
}

var AllBookArgs = `title, ISBN, author, publisher, stock, available, removeinfo`
var AllUserArgs = `id, name, password, overdue, type`
var AllRecordArgs = `record_id, book_id, user_id, IsReturned, borrow_date, return_date, deadline, extendtimes`

// BooksRowsScan : extract argvs from Query to struct Books
func (lib *Library) BooksRowsScan(rows *sql.Rows) (Books, error) {
	var res Books
	err := rows.Scan(&res.Title, &res.ISBN, &res.Author, &res.Publisher, &res.Stock, &res.Available, &res.RemoveInfo)
	return res, err
}

// BooksRowScan : extract argvs from QueryRow to struct Books
func (lib *Library) BooksRowScan(row *sql.Row) (Books, error) {
	var res Books
	err := row.Scan(&res.Title, &res.ISBN, &res.Author, &res.Publisher, &res.Stock, &res.Available, &res.RemoveInfo)
	return res, err
}

// UsersRowScan : extract argvs from QueryRow to struct Users
func (lib *Library) UsersRowScan(row *sql.Row) (Users, error) {
	var res Users
	err := row.Scan(&res.ID, &res.Name, &res.Password, &res.Overdue, &res.Type)
	return res, err
}

// RecordsRowsScan : extract argvs from Query to struct Records
func (lib *Library) RecordsRowsScan(rows *sql.Rows) (Records, error) {
	var res Records
	err := rows.Scan(&res.recordID, &res.bookID, &res.userID, &res.IsReturned, &res.borrowDate, &res.returnDate, &res.deadline, &res.extendTimes)
	return res, err
}

// RecordsRowScan : extract argvs from QueryRow to struct Records
func (lib *Library) RecordsRowScan(row *sql.Row) (Records, error) {
	var res Records
	err := row.Scan(&res.recordID, &res.bookID, &res.userID, &res.IsReturned, &res.borrowDate, &res.returnDate, &res.deadline, &res.extendTimes)
	return res, err
}

// CheckUserExists : check whether the user exists
func (lib *Library) CheckUserExists(userid string) error {
	var id string
	err := lib.db.QueryRow(`SELECT id FROM Userlist WHERE id = ?`, userid).Scan(&id)
	if err == sql.ErrNoRows {
		return ErrUserNotExists
	}
	return err
}

// CheckBookISBN : check whether the ID is legal
func (lib *Library) CheckBookISBN(err error) error {
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrBookNotExists
		}
		return err
	}
	return nil
}

// CheckBookExists : check whether the book is still in stock
func (lib *Library) CheckBookExists(ISBN string) error {
	var stock int
	err := lib.db.QueryRow(`SELECT stock FROM Booklist WHERE ISBN = ?`, ISBN).Scan(&stock)
	if err == sql.ErrNoRows {
		return ErrBookNotExists
	}
	if err == nil && stock <= 0 {
		err = ErrBookNotExists
	}
	return err
}

// CheckRecordID : check whether the ID is legal
func (lib *Library) CheckRecordID(err error) error {
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotBorrowed
		}
		return err
	}
	return nil
}

// CreateTables : created the tables in MySQL
func (lib *Library) CreateTables() error {
	sql := `CREATE TABLE IF NOT EXISTS Booklist(
				title VARCHAR(256) NOT NULL,
				ISBN VARCHAR(16) UNIQUE PRIMARY KEY,
				author VARCHAR(256) NOT NULL,
				publisher VARCHAR(256) NOT NULL,
				stock INT NOT NULL,
				available INT NOT NULL,
				removeInfo TEXT,
				CHECK (stock >= available AND stock >= 0)
			)`
	_, err := lib.db.Exec(sql)
	if err != nil {
		log.Println(err)
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS Userlist(
			id VARCHAR(16) PRIMARY KEY,
			name VARCHAR(256) NOT NULL,
			password VARCHAR(256) NOT NULL,
			overdue INT NOT NULL DEFAULT 0,
			type INT NOT NULL
			)`
	_, err = lib.db.Exec(sql)
	if err != nil {
		log.Println(err)
		return err
	}

	sql = `CREATE TABLE IF NOT EXISTS Recordlist(
		record_id INT PRIMARY KEY NOT NULL AUTO_INCREMENT,
		book_id VARCHAR(16) NOT NULL,
		user_id VARCHAR(16) NOT NULL,
		IsReturned BOOLEAN NOT NULL DEFAULT FALSE,
		borrow_date DATETIME NOT NULL,
		return_date DATETIME,
		deadline DATETIME NOT NULL,
		extendtimes INT NOT NULL,
		FOREIGN KEY (book_id) REFERENCES Booklist(ISBN),
		FOREIGN KEY (user_id) REFERENCES Userlist(id),
		CHECK (deadline >= borrow_date)
	)AUTO_INCREMENT=1`
	_, err = lib.db.Exec(sql)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// AddUser : add a user into the userlist
func (lib *Library) AddUser(user Users) error {
	_, err := lib.db.Exec(`INSERT INTO Userlist(id, name, password, type, overdue)
							VALUES (?, ?, ?, ?, ?)`,
		user.ID, user.Name, user.Password, user.Type, user.Overdue)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil

}

// IdentifyUser : to identify the user by the id and password
func (lib *Library) IdentifyUser(userid, password string) (Users, error) {
	var user Users
	user, err := lib.UsersRowScan(lib.db.QueryRow(`SELECT `+AllUserArgs+` FROM Userlist WHERE id = ?`, userid))

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(ErrPassword)
			return user, ErrPassword
		}
		log.Println(err)
		return user, err
	}

	if strings.Compare(user.Password, password) != 0 {
		err = ErrPassword
	}

	if err != nil {
		log.Println(err)
	}

	return user, err
}

// ModifyPassword : to modify user's password
func (lib *Library) ModifyPassword(userid, password string) error {
	_, err := lib.db.Exec(`UPDATE Userlist SET password = ? WHERE id = ?`, password, userid)
	if err != nil {
		log.Println(err)
	}

	return err
}

// AddBook : add a book into the library
func (lib *Library) AddBook(bookTitle, bookISBN, bookAuthor, bookPublisher string, bookStock int) (int, error) {
	var stock int
	row := lib.db.QueryRow(`SELECT `+`stock`+` FROM Booklist WHERE ISBN = ?;`, bookISBN)
	err := row.Scan(&stock)

	if err != nil {
		if err == sql.ErrNoRows {
			_, err = lib.db.Exec(`INSERT INTO Booklist(title, ISBN, author, publisher, stock, available)
						 VALUES (?, ?, ?, ?, ?, ?);`,
				bookTitle, bookISBN, bookAuthor, bookPublisher, bookStock, bookStock)
			if err != nil {
				log.Println("Insert Error: ", err)
				return -1, err
			}
			stock = bookStock
		} else {
			log.Println("QueryRow Error: ", err)
			return -1, err
		}
	} else {
		_, err = lib.db.Exec(`UPDATE Booklist
						 SET stock = stock + ?, available = available + ?
						 WHERE ISBN = ?;`,
			bookStock, bookStock, bookISBN)
		if err != nil {
			log.Println("Update Error: ", err)
			return -1, err
		}
		stock = stock + bookStock
	}

	log.Println("Added successfully.")
	return stock, nil
}

// RemoveBook : remove a book from the library
// if an admin lost the book, he or she can just run this function and add an info
// if a student lost the book, the borrow record must be modified before remove it
// require book's ISBN and the remove reason
func (lib *Library) RemoveBook(bookISBN, bookRemoveInfo string) (int, error) {
	var stock int
	row := lib.db.QueryRow(`SELECT `+`stock`+` FROM Booklist WHERE ISBN = ?`, bookISBN)
	err := row.Scan(&stock)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(ErrBookNotExists, " Operation failed.")
			return -1, ErrBookNotExists
		}
		log.Println("QueryErr: ", err)
		return -1, err
	} else {
		if stock == 0 {
			log.Println(ErrAllRemoved, " Operation failed.")
			return -1, ErrAllRemoved
		}
		_, err := lib.db.Exec(`UPDATE Booklist
						 SET stock = stock - 1, available = available - 1, removeinfo = CONCAT(?, removeInfo)
						 WHERE ISBN = ?;`, bookRemoveInfo, bookISBN)
		if err != nil {
			log.Println("The book exist. update error: ", err)
			return -1, err
		}
		stock = stock - 1
	}

	log.Println("Removed successfully.")
	return stock, nil
}

// QueryBookTitle : query books by title
func (lib *Library) QueryBookTitle(keyTitle string) ([]Books, error) {
	rows, err := lib.db.Query(`SELECT `+AllBookArgs+
		` FROM Booklist WHERE title LIKE ? 
		ORDER BY LENGTH(title) ASC, title ASC;`,
		"%"+keyTitle+"%")
	if err != nil {
		log.Println("Exec error: ", err)
		return nil, err
	}
	defer rows.Close()

	BookList := []Books{}
	var res Books
	for rows.Next() {
		res, err = lib.BooksRowsScan(rows)
		if err != nil {
			log.Println("rows.scan error: ", err)
			return nil, err
		}
		BookList = append(BookList, res)
	}

	return BookList, nil
}

// QueryBookAuthor : query books by author
func (lib *Library) QueryBookAuthor(keyAuthor string) ([]Books, error) {
	rows, err := lib.db.Query(`SELECT `+AllBookArgs+` FROM Booklist WHERE Author like ?;`, "%"+keyAuthor+"%")
	if err != nil {
		log.Println("Exec error: ", err)
		return nil, err
	}
	defer rows.Close()

	BookList := []Books{}
	var res Books
	for rows.Next() {
		res, err = lib.BooksRowsScan(rows)
		if err != nil {
			log.Println("rows.scan error: ", err)
			return nil, err
		}
		BookList = append(BookList, res)
	}

	return BookList, nil
}

// QueryBookISBN : query books by ISBN
func (lib *Library) QueryBookISBN(keyISBN string) ([]Books, error) {
	var res Books
	row := lib.db.QueryRow(`SELECT `+AllBookArgs+` FROM Booklist WHERE ISBN = ? AND stock > 0;`, keyISBN)
	res, err := lib.BooksRowScan(row)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(ErrBookNotExists)
			return nil, ErrBookNotExists
		}
		fmt.Println("Exec error: ", err)
		return nil, err
	}

	BookList := []Books{}

	if err != nil {
		log.Println("rows.scan error: ", err)
		return nil, err
	}
	BookList = append(BookList, res)

	return BookList, nil
}

// BorrowBook : borrow a book from the library
// borrow one book at a time
// book need to be returned in one month unless extended
// require book's ISBN, user's ID, and borrowDate
func (lib *Library) BorrowBook(bookISBN, userID string, borrowDate time.Time) error {
	var recordID int
	row := lib.db.QueryRow(`SELECT record_id FROM Recordlist WHERE book_id = ? AND user_id = ? AND IsReturned = 0`,
		bookISBN, userID)
	err := row.Scan(&recordID)
	if err != nil && err != sql.ErrNoRows {
		log.Println("check whether borrowed", err)
		return err
	}
	if err == nil {
		log.Println(ErrAlreadyBorrowed)
		return ErrAlreadyBorrowed
	}

	var availableAmount, stock int
	row = lib.db.QueryRow(`SELECT stock, available FROM Booklist WHERE ISBN = ? AND stock > 0`, bookISBN)
	err = lib.CheckBookISBN(row.Scan(&stock, &availableAmount))
	if err != nil {
		log.Println(err)
		return err
	}

	if stock <= 0 {
		log.Println(ErrBookNotExists)
		return ErrBookNotExists
	}
	if availableAmount <= 0 {
		log.Println(ErrBookNotAvailable)
		return ErrBookNotAvailable
	}

	_, err = lib.db.Exec(`UPDATE Booklist SET available = available - 1 WHERE ISBN = ?`, bookISBN)
	if err != nil {
		log.Println("Modify available: ", err)
		return err
	}

	deadline := borrowDate.AddDate(0, 1, 0)

	_, err = lib.db.Exec(`Insert INTO Recordlist (book_id, user_id, IsReturned, borrow_date, deadline, extendtimes)
							  VALUES (?, ?, 0, ?, ?, 0)`,
		bookISBN, userID, borrowDate, deadline)

	if err != nil {
		log.Println("Insert record: ", err)
		return err
	}

	log.Println("Borrowed successfully.")
	return nil
}

// CheckDeadline : check the deadline of returning of a borrowed book for students
// require book's ISBN, user's ID
func (lib *Library) CheckDeadline(bookISBN, userID string) error {
	var res Records
	err := lib.CheckBookExists(bookISBN)
	if err != nil {
		log.Println(err)
		return err
	}

	row := lib.db.QueryRow(`SELECT `+AllRecordArgs+` FROM Recordlist WHERE book_id = ? AND user_id = ? AND IsReturned = FALSE`, bookISBN, userID)
	res, err = lib.RecordsRowScan(row)
	err = lib.CheckRecordID(err)
	if err != nil {
		return err
	}

	if err != nil {
		fmt.Println("rows.scan error: ", err)
		return err
	}

	fmt.Println("Deadline: ", res.deadline.Format(timeTemplate))
	return nil
}

// CheckBorrowHistory : check the student's borrow history
// require user's ID
func (lib *Library) CheckBorrowHistory(userID string) ([]Records, error) {
	rows, err := lib.db.Query(`SELECT `+AllRecordArgs+` FROM Recordlist WHERE user_id = ? ORDER BY borrow_date DESC`, userID)
	if err != nil {
		fmt.Println("Record Error: ", err)
		return nil, err
	}
	defer rows.Close()

	RecordList := []Records{}
	var res Records

	for rows.Next() {
		res, err = lib.RecordsRowsScan(rows)
		if err != nil {
			fmt.Println("rows.scan error: ", err)
			return nil, err
		}
		RecordList = append(RecordList, res)
	}

	return RecordList, nil
}

// CheckUnreturned : check the student's unreturned books
// require user's ID
func (lib *Library) CheckUnreturned(userID string) ([]Records, error) {
	rows, err := lib.db.Query(`SELECT `+AllRecordArgs+` FROM Recordlist WHERE user_id = ? AND IsReturned = 0 ORDER BY borrow_date DESC`, userID)
	if err != nil {
		fmt.Println("Record Error: ", err)
		return nil, err
	}
	defer rows.Close()

	RecordList := []Records{}
	var res Records

	for rows.Next() {
		res, err = lib.RecordsRowsScan(rows)
		if err != nil {
			fmt.Println("rows.scan error:", err)
			return nil, err
		}
		RecordList = append(RecordList, res)
	}

	return RecordList, nil
}

// CheckOverdue : check if a given student has any overdue
// require user's ID
func (lib *Library) CheckOverdue(userID string, now time.Time) (int, []Records, error) {
	rows, err := lib.db.Query(`SELECT `+AllRecordArgs+` FROM Recordlist WHERE user_id = ? AND IsReturned = FALSE`, userID)
	if err != nil {
		log.Println("Record Error: ", err)
		return -1, nil, err
	}
	defer rows.Close()

	var IsExtended bool
	var overdue = 0
	var RecordList []Records

	for rows.Next() {
		res, err := lib.RecordsRowsScan(rows)
		if err != nil {
			log.Println("rows.scan error: ", err)
			return -1, nil, err
		}

		IsExtended = false
		for res.extendTimes < 3 {
			if now.After(res.deadline) {
				res.extendTimes = res.extendTimes + 1
				res.deadline.AddDate(0, 1, 0)
				IsExtended = true
			} else {
				break
			}
		}

		if IsExtended {
			_, err = lib.db.Exec(`UPDATE Recordlist
									SET extendtimes = ?, deadline = ?
									WHERE record_id = ?`,
				res.extendTimes, res.deadline, res.recordID)
			if err != nil {
				log.Println("Update Error: ", err)
				return -1, nil, err
			}
		}

		if now.After(res.deadline) {
			overdue = overdue + 1
			RecordList = append(RecordList, res)
		}
	}

	_, err = lib.db.Exec(`UPDATE Userlist SET overdue = ? WHERE id = ?`, overdue, userID)
	if err != nil {
		log.Println("Update Error: ", err)
		return -1, nil, err
	}

	return overdue, RecordList, nil
}

// ReturnBook : return a borrowed book
// require book's ISBN and user's ID
func (lib *Library) ReturnBook(bookISBN, userID string) error {
	err := lib.CheckBookExists(bookISBN)
	if err != nil {
		log.Println(err)
		return err
	}

	var ddl time.Time
	var recordID string
	row := lib.db.QueryRow(`SELECT `+`record_id, deadline`+` FROM Recordlist WHERE user_id = ? AND book_id = ? AND IsReturned = 0`, userID, bookISBN)
	err = row.Scan(&recordID, &ddl)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(ErrNotBorrowed)
			return ErrNotBorrowed
		}
		log.Println("QueryRow: ", err)
		return err
	}

	var flag = 0
	var now = time.Now()
	if now.After(ddl) {
		flag = 1
	}

	_, err = lib.db.Exec(`UPDATE Recordlist SET IsReturned = 1, return_date = ? WHERE record_id = ?`, now, recordID)
	if err != nil {
		log.Println("B", err)
		return err
	}

	_, err = lib.db.Exec(`UPDATE Userlist SET overdue = overdue - ? WHERE id = ?`, flag, userID)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = lib.db.Exec(`UPDATE Booklist SET available = available + 1 WHERE ISBN = ?`, bookISBN)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Returned successfully.")
	return nil
}

// ExtendDeadline - extend deadline of a borrowed book for a given user
// require book_id, user_id
func (lib *Library) ExtendDeadline(bookISBN, userID string) error {
	err := lib.CheckBookExists(bookISBN)
	if err != nil {
		log.Println(err)
		return err
	}

	var ddl time.Time
	var extended, recordID int

	row := lib.db.QueryRow(`SELECT `+`record_id, deadline, extendtimes`+` FROM Recordlist WHERE book_id = ? AND user_id = ? AND IsReturned = FALSE`, bookISBN, userID)
	err = lib.CheckRecordID(row.Scan(&recordID, &ddl, &extended))

	if err != nil {
		log.Println(err)
		return err
	}

	if extended >= 3 {
		log.Println(ErrNoMoreExtended)
		return ErrNoMoreExtended
	}

	ddl.AddDate(0, 1, 0)
	_, err = lib.db.Exec(`UPDATE Recordlist
							SET extendtimes = extendtimes + 1, deadline = ?
							WHERE record_id = ?`,
		ddl, recordID)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Extended successfully.")
	return nil
}

// The above codes are about operating mysql and have test functions
/*-------------------------------------------------------*/
// Following codes are about interacting with the terminal and don't have test function

// GetInputString : to ensure that user not input an empty string
func (lib *Library) GetInputString(field string) string {
	var ret string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(field)
	for scanner.Scan() {
		ret = scanner.Text()
		if ret != "" {
			break
		}
		fmt.Print(field)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	return ret
}

// Register : for new user to register
// once they type in their username, the program will check whether the id is available immediately and ask for another try if needed
func (lib *Library) Register(auth int) {
	var user Users
	for true {
		user.ID = lib.GetInputString("Username: ")
		err := lib.CheckUserExists(user.ID)
		if err == nil {
			fmt.Println("Sorry. The username have already been registered.")
			return
		} else {
			if err == ErrUserNotExists {
				break
			}
			log.Println(err)
			return
		}
	}
	user.Name = lib.GetInputString("RealName: ")
	user.Password = lib.GetInputString("Password: ")
	confirmPassword := lib.GetInputString("ConfirmPassword: ")
	if user.Password != confirmPassword {
		log.Println("Password and ConfirmPassword don't match.")
		return
	}
	user.Type = 1
	if auth == 0 {
		fmt.Println("Please decide the user mode. Type 0 for administrator, 1 for normal user")
		for true {
			fmt.Print("UserMode: ")
			var mode string
			fmt.Scanln(&mode)
			if mode == "0" {
				user.Type = 0
				break
			} else if mode == "1" {
				user.Type = 1
				break
			}
			fmt.Println(mode, ": user mode not found")
		}

	}
	if err := lib.AddUser(user); err == nil {
		log.Println("Registered Successfully. You can login now.")
	}
}

// PrintBookQuery : to print book information
func (lib *Library) PrintBookQuery(books []Books, usermode int) {
	if usermode == 0 {
		if len(books) == 0 {
			log.Println(ErrBookNotExists)
			return
		}
		t := table.Table(books)
		fmt.Println(t)
	} else {
		type Data struct {
			ISBN, Title, Author, Publisher string
			Stock, Available               int
		}
		var res []Data
		for _, now := range books {
			if now.Stock > 0 {
				tmp := Data{now.ISBN, now.Title, now.Author, now.Publisher, now.Stock, now.Available}
				res = append(res, tmp)
			}
		}
		if len(res) == 0 {
			log.Println(ErrBookNotExists)
		} else {
			t := table.Table(res)
			fmt.Println(t)
		}
	}
}

// PrintOverdue : to print the users' overdue information
func (lib *Library) PrintOverdue(overdue int, records []Records) {
	if overdue > 0 {
		fmt.Println("Warning: You've got overdue(s). Please turn the book(s) back ASAP.")
	}
	if overdue > 3 {
		fmt.Println("Warning: ", ErrUserSuspended)
	}
	fmt.Println("overdue: ", overdue)
	if overdue > 0 {
		lib.PrintUnreturned(records)
	}
}

// PrintUnreturned : print users' unreturned list with deadline
func (lib *Library) PrintUnreturned(records []Records) {
	type data struct {
		recordID    string
		ISBN        string
		Title       string
		ExtendTimes int
		BorrowDate  string
		Deadline    string
	}
	var res []data
	for _, now := range records {
		Title, _ := lib.QueryBookISBN(now.bookID)
		res = append(res, data{now.recordID, now.bookID, Title[0].Title, now.extendTimes, now.borrowDate.Format(timeTemplate), now.deadline.Format(timeTemplate)})
	}

	if len(res) != 0 {
		t := table.Table(res)
		fmt.Println(t)
	} else {
		fmt.Println("No unreturned book.")
	}
}

// PrintHistory : print users' history with return date
func (lib *Library) PrintHistory(records []Records, sign error) {
	if sign != nil {
		return
	}
	type data struct {
		recordID    string
		ISBN        string
		Title       string
		IsReturned  bool
		ExtendTimes int
		BorrowDate  string
		ReturnDate  string
	}
	var ss []data
	for _, now := range records {
		Title, _ := lib.QueryBookISBN(now.bookID)
		var tmp string
		if !now.returnDate.Valid {
			tmp = "NULL"
		} else {
			tmp = now.returnDate.Time.Format(timeTemplate)
		}
		ss = append(ss, data{now.recordID, now.bookID, Title[0].Title, now.IsReturned, now.extendTimes, now.borrowDate.Format(timeTemplate), tmp})
	}

	t := table.Table(ss)
	fmt.Println(t)
}

// Servertime : to serve the user
// users can read readme file or type `help` for detailed instructions
func (lib *Library) Servetime(user Users) {
	var input string
	var book Books

	var userID string

	if user.Type == 1 {
		userID = user.ID
	}

	for true {
		fmt.Print(user.Name, "@library: ")
		input = lib.GetInputString("")
		if input == "exit" {
			return
		} else if input == "help" {
			fmt.Println(help)
		} else if input == "title" {
			book.Title = lib.GetInputString("BookTitle: ")
			res, err := lib.QueryBookTitle(book.Title)
			if err == nil {
				lib.PrintBookQuery(res, user.Type)
			}
		} else if input == "author" {
			book.Author = lib.GetInputString("BookAuthor: ")
			res, err := lib.QueryBookAuthor(book.Author)
			if err == nil {
				lib.PrintBookQuery(res, user.Type)
			}
		} else if input == "isbn" {
			book.ISBN = lib.GetInputString("BookISBN: ")
			res, err := lib.QueryBookISBN(book.ISBN)
			if err == nil {
				lib.PrintBookQuery(res, user.Type)
			}
		} else if user.Type > 1 {
			fmt.Println(input, ": command not found")
		} else if input == "borrow" {
			if user.Type < 1 {
				userID = lib.GetInputString("Username: ")
				err := lib.CheckUserExists(userID)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
			overdue, recordlist, err := lib.CheckOverdue(userID, time.Now())
			if err == nil {
				if overdue > 0 {
					lib.PrintOverdue(overdue, recordlist)
				}
				if overdue <= 3 {
					book.ISBN = lib.GetInputString("BookISBN: ")
					lib.BorrowBook(book.ISBN, userID, time.Now())
				}
			}
		} else if input == "return" {
			if user.Type < 1 {
				userID = lib.GetInputString("Username: ")
				err := lib.CheckUserExists(userID)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
			book.ISBN = lib.GetInputString("BookISBN: ")
			lib.ReturnBook(book.ISBN, userID)
		} else if input == "deadline" {
			if user.Type < 1 {
				userID = lib.GetInputString("Username: ")
				err := lib.CheckUserExists(userID)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
			book.ISBN = lib.GetInputString("BookISBN: ")
			lib.CheckOverdue(userID, time.Now())
			lib.CheckDeadline(book.ISBN, userID)
		} else if input == "extend" {
			if user.Type < 1 {
				userID = lib.GetInputString("Username: ")
				err := lib.CheckUserExists(userID)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
			book.ISBN = lib.GetInputString("BookISBN: ")
			lib.CheckOverdue(userID, time.Now())
			lib.ExtendDeadline(book.ISBN, userID)
		} else if input == "history" {
			if user.Type < 1 {
				userID = lib.GetInputString("Username: ")
				err := lib.CheckUserExists(userID)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
			lib.CheckOverdue(userID, time.Now())
			lib.PrintHistory(lib.CheckBorrowHistory(userID))
		} else if input == "unreturned" {
			if user.Type < 1 {
				userID = lib.GetInputString("Username: ")
				err := lib.CheckUserExists(userID)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
			lib.CheckOverdue(userID, time.Now())
			res, _ := lib.CheckUnreturned(userID)
			lib.PrintUnreturned(res)
		} else if input == "overdue" {
			if user.Type < 1 {
				userID = lib.GetInputString("Username: ")
				err := lib.CheckUserExists(userID)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
			overdue, record, _ := lib.CheckOverdue(userID, time.Now())
			lib.PrintOverdue(overdue, record)
		} else if input == "pw" {
			password := lib.GetInputString("Password: ")
			if password != user.Password {
				log.Println(ErrPassword)
			} else {
				password = lib.GetInputString("NewPassword: ")
				confirmpw := lib.GetInputString("ConfirmNewPassword: ")
				if password == confirmpw {
					if lib.ModifyPassword(user.ID, password) != nil {
						user.Password = password
					}
				}
			}
		} else if user.Type > 0 {
			fmt.Println(input, ": command not found")
		} else if input == "adduser" {
			lib.Register(0)
		} else if input == "addbook" {
			book.ISBN = lib.GetInputString("BookISBN: ")
			fmt.Print("BookTitle: ")
			fmt.Scanln(&book.Title)
			fmt.Print("BookAuthor: ")
			fmt.Scanln(&book.Author)
			fmt.Print("BookPublisher: ")
			fmt.Scanln(&book.Publisher)
			fmt.Print("BookStock: ")
			fmt.Scanln(&book.Stock)
			lib.AddBook(book.Title, book.ISBN, book.Author, book.Publisher, book.Stock)
		} else if input == "removebook" {
			book.ISBN = lib.GetInputString("BookISBN: ")
			book.RemoveInfo.String = lib.GetInputString("RemoveInfo: ")
			book.RemoveInfo.String += fmt.Sprintf("Removed by %s at %s", user.ID, time.Now().Format(timeTemplate))
			lib.RemoveBook(book.ISBN, book.RemoveInfo.String)
		} else if input == "userpw" {
			username := lib.GetInputString("Username: ")
			password := lib.GetInputString("NewPassword: ")
			confirmpw := lib.GetInputString("ConfirmNewPassword: ")
			if password == confirmpw {
				lib.ModifyPassword(username, password)
			}
		} else {
			fmt.Println(input, ": command not found")
		}
	}
}

func main() {
	fmt.Println("Welcome to the Library Management System!")
	fmt.Println("Type \"help\" for more information.")
	s, _ := ioutil.ReadFile("readme.txt")
	help = string(s)

	var lib Library
	var input string

	lib.ConnectDB()
	lib.CreateTables()

	for true {
		fmt.Print(">> ")
		input = lib.GetInputString("")
		if input == "exit" {
			break
		} else if input == "help" {
			fmt.Println(help)
		} else if input == "login" {
			var username, password string
			username = lib.GetInputString("Username: ")
			password = lib.GetInputString("Password: ")
			user, err := lib.IdentifyUser(username, password)
			if err == nil {
				log.Println("Login Successfully.")
				lib.Servetime(user)
			}
		} else if input == "guest-mode" {
			var user Users
			user.Name = "guest"
			user.Type = 2
			lib.Servetime(user)
		} else if input == "register" {
			lib.Register(-1)
		} else if input != "" {
			fmt.Println(input, ": command not found")
			fmt.Println("Type \"help\" for more information.")
		}
	}
}
