"exit" -- exit the management system/log out the system
"help" -- ask for instructions

"register" -- to register a new account
"guest-mode" -- to log in the system as a guest
"login" -- to login with your account

since you log into the system, there are three user modes: guests, normal readers and administrators

for guests:
	"title" -- to query book(s) by title
	"author" -- to query book(s) by author
	"isbn" -- to query book(s) by ISBN

for normal readers:
	they can do all the operations mentioned above, and following extra operations
	"pw" -- to reset one's own password
	"borrow" -- to borrow a book
	"return" -- to return a book
	"extend" -- to extend the deadline of returning a book
	"deadline" -- to query the deadline of a borrowed book
	"overdue" -- to query the amount of overdue books
	"unreturned" -- to view all the unreturned books
	"history" -- to view all the borrow history

for administrators:
	they can do all the operations mentioned above, and following extra operations
	"adduser" -- add a new user and set the user mode
	"addbook" -- add book
	"userpw" -- help user who forgot his/her password to reset it, 
		    and please be very cautious when doing this operation
	"removebook" -- remove book and add remove information
			// when remove a book, if it's about a student lost it,
			// make sure you've fine the student and got the book "returned",
			// or it may have impact on the whole system
