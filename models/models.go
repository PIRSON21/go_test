package models

import "database/sql"

// Структура для отображения данных о теме
type Theme struct {
	ThemeID           int
	Title             string
	TeacherID         int
	TeacherFirstName  string
	TeacherMiddleName sql.NullString
	TeacherLastName   string
	StudentID         sql.NullInt64
	StudentFirstName  sql.NullString
	StudentLastName   sql.NullString
	StudentRecordBook sql.NullString
	TestMark          sql.NullInt64
	DiplomMark        sql.NullInt64
}

// Структура для отображения данных о студенте
type Student struct {
	StudentID  int
	FirstName  string
	LastName   string
	RecordBook string
	FacultyID  int
	GroupID    int
	ThemeTitle string
	TestMark   sql.NullInt64
	DiplomMark sql.NullInt64
}

// Структура для отображения данных о преподавателе
type Teacher struct {
	TeacherID     int
	FirstName     string
	MiddleName    sql.NullString
	LastName      string
	Grade         *Grade
	AcademicTitle *AcademicTitle
	Departments   []Department
	Phone         sql.NullString
	Email         sql.NullString
}

type Grade struct {
	GradeID   int
	GradeName string
}

type AcademicTitle struct {
	AcademicTitleID   sql.NullInt64
	AcademicTitleName sql.NullString
}

type Department struct {
	DepartmentID   int
	DepartmentName string
}

type Group struct {
	GroupID   int
	GroupName string
}

type Faculty struct {
	FacultyID   int
	FacultyName string
}
