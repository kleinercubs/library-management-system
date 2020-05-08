package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

var lib = Library{}

func TestCreateTables(t *testing.T) {
	lib.ConnectDB()
	_, err := lib.db.Exec(`DROP TABLE Recordlist`)
	if err != nil {
		panic(err)
	}
	_, err = lib.db.Exec(`DROP TABLE Booklist`)
	if err != nil {
		panic(err)
	}
	_, err = lib.db.Exec(`DROP TABLE Userlist;`)
	if err != nil {
		panic(err)
	}

	err = lib.CreateTables()
	if err != nil {
		t.Errorf("can't create tables")
	}
}

func TestAddUsers(t *testing.T) {
	var tests = []struct {
		/*	ID, Name, Password string
			Overdue, Type      int*/
		testid int
		user   Users
		err    error
	}{
		{0, Users{`root`, `admin`, `axd`, 0, 0}, nil},
		{1, Users{`18307130006`, `Alicia`, `578152`, 0, 1}, nil},
		{2, Users{`18307130068`, `Brandon`, `987430`, 0, 1}, nil},
		{3, Users{`18307130078`, `Cary`, `15652`, 0, 1}, nil},
		{4, Users{`18307130091`, `Doreen`, `740509`, 0, 1}, nil},
		{5, Users{`18307130083`, `Eric`, `194687`, 0, 1}, nil},
		{6, Users{`18307130002`, `Florrick`, `979148`, 0, 1}, nil},
		{7, Users{`18307130018`, `Grace`, `979136`, 0, 1}, nil},
		{8, Users{`18307130090`, `Helen`, `675966`, 0, 1}, nil},
		{9, Users{`18307130012`, `Irene`, `635982`, 0, 1}, nil},
		{10, Users{`18307130041`, `Jack`, `736437`, 0, 1}, nil},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.testid)
		t.Run(testname, func(t *testing.T) {
			err := lib.AddUser(tt.user)
			if err != tt.err {
				t.Errorf("got %v, want nil", err)
			}
		})
	}
}

func TestModifyPassword(t *testing.T) {
	var tests = []struct {
		testid           int
		userID, password string
		err              error
	}{
		{0, `root`, `root`, nil},
		{1, `18307130068`, `asdfa`, nil},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.testid)
		t.Run(testname, func(t *testing.T) {
			err := lib.ModifyPassword(tt.userID, tt.password)
			if err != tt.err {
				t.Errorf("got %v, want nil", err)
			}
		})
	}
}

func TestCheckUserExists(t *testing.T) {
	var tests = []struct {
		testid int
		userID string
		err    error
	}{
		{0, `111`, ErrUserNotExists},
		{1, `18307130006`, nil},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.testid)
		t.Run(testname, func(t *testing.T) {
			err := lib.CheckUserExists(tt.userID)
			if err != tt.err {
				t.Errorf("got %v, want nil", err)
			}
		})
	}
}

func TestIdentifyUser(t *testing.T) {
	var tests = []struct {
		testid   int
		userid   string
		password string
		Type     int
		err      error
	}{
		{0, `18307130006`, `578152`, 1, nil},
		{1, `18307130006`, `123456`, 1, ErrPassword},
		{2, `root`, `root`, 0, nil},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.testid)
		t.Run(testname, func(t *testing.T) {
			user, err := lib.IdentifyUser(tt.userid, tt.password)
			if err != tt.err {
				t.Errorf("got %v, want %v", err, tt.err)
			}
			if err == nil && user.Type != tt.Type {
				t.Errorf("got %v, want %v", user.Type, tt.Type)
			}
		})
	}
}

func TestAddBook(t *testing.T) {
	var tests = []struct {
		ISBN, Title, Author, Publisher string
		Stock                          int
		res                            int
	}{
		{`978-0735219090`, `Where the Crawdads Sing`, `Delia Owens`, `G.P. Putnam's Sons; Later Printing edition (August 14, 2018)`, 3, 3},
		{`978-1984801258`, `Untamed`, `Glennon Doyle`, `The Dial Press (March 10, 2020)`, 2, 2},
		{`978-1338635171`, `The Ballad of Songbirds and Snakes (A Hunger Games Novel)`, `Suzanne Collins`, `Scholastic Press (May 19, 2020)`, 1, 1},
		{`978-0062820181`, `Magnolia Table, Volume 2: A Collection of Recipes for Gathering`, `Joanna Gaines`, `William Morrow Cookbooks (April 7, 2020)`, 3, 3},
		{`978-1984822185`, `Normal People: A Novel`, `Sally Rooney`, `Hogarth; Reprint edition (February 18, 2020)`, 1, 1},
		{`978-0385545938`, `Camino Winds`, `John Grisham`, `Doubleday (April 28, 2020)`, 3, 3},
		{`978-0134123837`, `Computer Systems: A Programmer's Perspective (3rd Edition)`, `Randal E. Bryant, David R. O'Hallaron`, `Pearson; 3 edition (July 6, 2015)`, 1, 1},
		{`978-0262033848`, `Introduction to Algorithms, 3rd Edition`, `Thomas H. Cormen, Charles E. Leiserson, Ronald L. Rivest , Clifford Stein`, `The MIT Press; 3rd edition (July 31, 2009)`, 2, 2},
		{`978-0134123837`, `Computer Systems: A Programmer's Perspective (3rd Edition)`, ` Randal E. Bryant, David R. O'Hallaron`, `Pearson; 3 edition (July 6, 2015)`, 3, 4},
		{`978-0395680902`, `Eye of the Elephant Pa`, `Delia Owens`, `Mariner (October 29, 1993)`, 2, 2},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.ISBN)
		t.Run(testname, func(t *testing.T) {
			ans, err := lib.AddBook(tt.Title, tt.ISBN, tt.Author, tt.Publisher, tt.Stock)
			if ans != tt.res || err != nil {
				t.Errorf("got %v, want nil", err)
				t.Errorf("got %d, want %d", ans, tt.res)
			}
		})
	}
}

func TestCheckBookExists(t *testing.T) {
	var tests = []struct {
		testid int
		ISBN   string
		err    error
	}{
		{0, `978-0134123837`, nil},
		{1, `111`, ErrBookNotExists},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.testid)
		t.Run(testname, func(t *testing.T) {
			err := lib.CheckBookExists(tt.ISBN)
			if err != tt.err {
				t.Errorf("got %v, want nil", err)
			}
		})
	}
}

func TestRemoveBook(t *testing.T) {
	var tests = []struct {
		testid int
		ISBN   string
		Info   string
		res    int
		err    error
	}{
		{0, `978-0735219090`, `lost`, 2, nil},
		{1, `978-1984801258`, `renew`, 1, nil},
		{2, `978-1338635171`, `lost`, 0, nil},
		{3, `978-0735219090`, `lost`, 1, nil},
		{4, `978-1338635171`, "", -1, ErrAllRemoved},
		{5, `978-0735219090`, "", 0, nil},
		{6, `234-2342135234`, `no reason`, -1, ErrBookNotExists},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.testid)
		t.Run(testname, func(t *testing.T) {
			ans, err := lib.RemoveBook(tt.ISBN, tt.Info)
			if ans != tt.res || err != tt.err {
				t.Errorf("got %d, want %d", ans, tt.res)
				t.Errorf("got %v, want %v", err, tt.err)
			}
		})
	}
}

func TestQueryBookTitle(t *testing.T) {
	var tests = []string{
		`camino winds`,
		`the`,
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt)
		t.Run(testname, func(t *testing.T) {
			ans, err := lib.QueryBookTitle(tt)
			if err != nil {
				t.Errorf("%v", err)
			}
			lower := strings.ToLower(tt)
			for _, now := range ans {
				if !strings.Contains(strings.ToLower(now.Title), lower) {
					t.Errorf("got %s, want %s", now.Title, tt)
				}
			}
		})
	}
}

func TestQueryBookAuthor(t *testing.T) {
	var tests = []string{
		`Glennon Doyle`,
		`delia owens`,
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt)
		t.Run(testname, func(t *testing.T) {
			ans, err := lib.QueryBookAuthor(tt)
			if err != nil {
				t.Errorf("%v", err)
			}
			lower := strings.ToLower(tt)
			for _, now := range ans {
				if !strings.Contains(strings.ToLower(now.Author), lower) {
					t.Errorf("got %s, want %s", now.Author, tt)
				}
			}
		})
	}
}

func TestQueryBookISBN(t *testing.T) {
	var tests = []struct {
		query string
		err   error
	}{
		{`123-1234567890`, ErrBookNotExists},
		{`978-0735219090`, ErrBookNotExists},
		{`978-0395680902`, nil},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.query)
		t.Run(testname, func(t *testing.T) {
			ans, err := lib.QueryBookISBN(tt.query)
			if err != tt.err {
				t.Errorf("got %v, want %v", err, tt.err)
			}
			for _, now := range ans {
				if strings.Compare(now.ISBN, tt.query) != 0 {
					t.Errorf("got ans, want res")
				}
			}
		})
	}
}

func TestBorrowBook(t *testing.T) {
	var tests = []struct {
		testid           int
		bookISBN, userID string
		borrowDate       time.Time
		err              error
	}{
		{0, `978-0262033848`, `18307130002`, time.Date(2019, time.November, 1, 14, 0, 0, 0, time.UTC), nil},
		{1, `978-0385545938`, `18307130002`, time.Date(2020, time.January, 9, 14, 0, 0, 0, time.UTC), nil},
		{2, `978-1984822185`, `18307130002`, time.Date(2020, time.January, 10, 14, 0, 0, 0, time.UTC), nil},
		{3, `978-0262033848`, `18307130002`, time.Date(2020, time.January, 14, 10, 0, 0, 0, time.UTC), ErrAlreadyBorrowed},
		{4, `978-0735219090`, `18307130012`, time.Date(2020, time.April, 1, 14, 0, 0, 0, time.UTC), ErrBookNotExists},
		{5, `978-0062820181`, `18307130018`, time.Date(2020, time.April, 1, 14, 0, 0, 0, time.UTC), nil},
		{6, `978-0134123837`, `18307130018`, time.Date(2020, time.April, 15, 14, 0, 0, 0, time.UTC), nil},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.testid)
		t.Run(testname, func(t *testing.T) {
			err := lib.BorrowBook(tt.bookISBN, tt.userID, tt.borrowDate)
			if err != tt.err {
				t.Errorf("got %v, want %v", err, tt.err)
			}
		})
	}
}

func TestCheckDeadline(t *testing.T) {
	var tests = []struct {
		testid           int
		bookISBN, userID string
		err              error
	}{
		{0, `978-0134123837`, `18307130018`, nil},
		{1, `978-1234567890`, `18307130018`, ErrBookNotExists},
		{2, `978-0385545938`, `18307130018`, ErrNotBorrowed},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.testid)
		t.Run(testname, func(t *testing.T) {
			err := lib.CheckDeadline(tt.bookISBN, tt.userID)
			if err != tt.err {
				t.Errorf("got %v, want nil", err)

			}
		})
	}
}

func TestCheckBorrowHistory(t *testing.T) {
	var tests = []struct {
		userID string
		err    error
	}{
		{`18307130018`, nil},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.userID)
		t.Run(testname, func(t *testing.T) {
			res, err := lib.CheckBorrowHistory(tt.userID)
			if err != nil {
				if err != tt.err {
					t.Errorf("got %v, want nil", err)
				}
			} else {
				for _, ans := range res {
					if strings.Compare(ans.userID, tt.userID) != 0 {
						t.Errorf("got %s, want %s", ans.userID, tt.userID)
					}
				}
			}
		})
	}
}

func TestCheckUnreturned(t *testing.T) {
	var tests = []struct {
		userID string
		err    error
	}{
		{`18307130018`, nil},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%s", tt.userID)
		t.Run(testname, func(t *testing.T) {
			res, err := lib.CheckUnreturned(tt.userID)
			if err != nil {
				if err != tt.err {
					t.Errorf("got %v, want nil", err)
				}
			} else {
				for _, ans := range res {
					if strings.Compare(ans.userID, tt.userID) != 0 || ans.IsReturned == true {
						t.Errorf("got %s, want %s", ans.userID, tt.userID)
					}
				}
			}
		})
	}
}

func TestCheckOverdue(t *testing.T) {
	var tests = []struct {
		testID           int
		bookISBN, userID string
		borrowDate       time.Time
		overdue          int
		err              error
	}{
		{0, `978-0062820181`, `18307130006`, time.Date(2019, time.November, 1, 14, 0, 0, 0, time.UTC), 1, nil},
		{1, `978-0134123837`, `18307130006`, time.Date(2020, time.January, 9, 14, 0, 0, 0, time.UTC), 2, nil},
		{2, `978-0262033848`, `18307130006`, time.Date(2020, time.January, 10, 14, 0, 0, 0, time.UTC), 3, nil},
		{3, `978-0385545938`, `18307130006`, time.Date(2020, time.January, 15, 14, 0, 0, 0, time.UTC), 4, nil},
		{4, `978-0395680902`, `18307130006`, time.Date(2020, time.April, 15, 14, 0, 0, 0, time.UTC), 4, nil},
	}

	var now = time.Now()
	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.testID)
		t.Run(testname, func(t *testing.T) {
			if tt.testID != 5 {
				_, err := lib.db.Exec(`INSERT INTO Recordlist(book_id, user_id, IsReturned, borrow_date, deadline, extendtimes)
						VALUES (?, ?, 0, ?, ?, 0)`,
					tt.bookISBN, tt.userID, tt.borrowDate, tt.borrowDate.AddDate(0, 1, 0))
				if err != nil {
					t.Errorf("exec err %v", err)
				}
				_, err = lib.db.Exec(`UPDATE Booklist SET available = available - 1 WHERE ISBN = ?`,
					tt.bookISBN)
				if err != nil {
					t.Errorf("exec err %v", err)
				}
			}
			res, _, err := lib.CheckOverdue(tt.userID, now)
			if err != nil {
				if err != tt.err {
					t.Errorf("got %v, want nil", err)
				}
			} else {
				if res != tt.overdue {
					t.Errorf("got %d, want %d", res, tt.overdue)
				}
			}
		})
	}
}

func TestExtendDeadline(t *testing.T) {
	var tests = []struct {
		testID           int
		bookISBN, userID string
		err              error
	}{
		{0, `123-0062820181`, `18307130006`, ErrBookNotExists},
		{1, `978-0385545938`, `18307130006`, ErrNoMoreExtended},
		{2, `978-0395680902`, `18307130006`, nil},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.testID)
		t.Run(testname, func(t *testing.T) {
			err := lib.ExtendDeadline(tt.bookISBN, tt.userID)

			if err != tt.err {
				t.Errorf("got %v, want nil", err)
			}

		})
	}
}

func TestReturnBook(t *testing.T) {
	var tests = []struct {
		testID           int
		bookISBN, userID string
		err              error
	}{
		{0, `123-0062820181`, `18307130006`, ErrBookNotExists},
		{1, `978-0385545938`, `18307130006`, nil},
		{2, `978-0395680902`, `18307130006`, nil},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.testID)
		t.Run(testname, func(t *testing.T) {
			err := lib.ReturnBook(tt.bookISBN, tt.userID)
			if err != tt.err {
				t.Errorf("got %v, want nil", err)
			}

		})
	}
}
