package db

import (
	"database/sql"
	"fmt"
	"github.com/PIRSON21/lab3/logging"
	"github.com/PIRSON21/lab3/models"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

var db *sql.DB

// Функция для инициализации базы данных
func InitDB() error {
	var err error
	db, err = sql.Open("mysql", "root:cAz6wv5fhJ0RCgqPVGMY@tcp(127.0.0.1:3306)/ibm")
	if err != nil {
		return err
	}

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return err
	}

	logging.Info("Подключение к базе данных установлено")
	return nil
}

// Функция для получения списка тем
func GetThemes(filter string) ([]models.Theme, error) {
	var rows *sql.Rows
	var err error
	if filter == "no-student" {
		rows, err = db.Query(`
		SELECT t.theme_id, t.theme_title, t.teacher_id, te.first_name, te.middle_name, te.last_name, NULL AS student_id, NULL AS first_name, NULL AS last_name, NULL AS test_mark, NULL AS diplom_mark
		FROM theme t
		LEFT JOIN teacher te ON t.teacher_id = te.teacher_id
		WHERE t.student_record_book IS NULL`)
		if err != nil {
			return nil, err
		}
	} else if filter == "no-grade" {
		rows, err = db.Query(`
		SELECT t.theme_id, t.theme_title, t.teacher_id, te.first_name, te.middle_name, te.last_name, s.student_id, s.first_name, s.last_name, m.test_mark, m.diplom_mark
		FROM theme t
		LEFT JOIN teacher te ON t.teacher_id = te.teacher_id
		LEFT JOIN student s ON t.student_record_book = s.student_record_book
		LEFT JOIN mark m ON s.student_record_book = m.student_record_book
		WHERE t.student_record_book IS NOT NULL AND (m.test_mark IS NULL OR m.diplom_mark IS NULL)`)
		if err != nil {
			return nil, err
		}
	} else {
		rows, err = db.Query(`
		SELECT t.theme_id, t.theme_title, t.teacher_id, te.first_name, te.middle_name, te.last_name, s.student_id, s.first_name, s.last_name, m.test_mark, m.diplom_mark
		FROM theme t
		LEFT JOIN teacher te ON t.teacher_id = te.teacher_id
		LEFT JOIN student s ON t.student_record_book = s.student_record_book
		LEFT JOIN mark m ON s.student_record_book = m.student_record_book`)
		if err != nil {
			return nil, err
		}
	}

	defer rows.Close()

	var themes []models.Theme
	for rows.Next() {
		var theme models.Theme
		if err := rows.Scan(&theme.ThemeID, &theme.Title, &theme.TeacherID, &theme.TeacherFirstName, &theme.TeacherMiddleName, &theme.TeacherLastName, &theme.StudentID, &theme.StudentFirstName, &theme.StudentLastName, &theme.TestMark, &theme.DiplomMark); err != nil {
			return nil, err
		}
		themes = append(themes, theme)
	}

	return themes, nil
}

// Закрытие подключения
func Close() {
	if err := db.Close(); err != nil {
		logging.Error("Ошибка при закрытии базы данных: ", err)
	}
}

func GetThemeToEdit(id string) (*models.Theme, error) {
	rows, err := db.Query(`
		SELECT t.theme_id, t.theme_title, t.teacher_id, te.first_name, te.middle_name, te.last_name
		FROM theme t
		LEFT JOIN teacher te ON t.teacher_id = te.teacher_id
		WHERE t.theme_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var theme models.Theme
	if rows.Next() {
		if err := rows.Scan(&theme.ThemeID, &theme.Title, &theme.TeacherID, &theme.TeacherFirstName, &theme.TeacherMiddleName, &theme.TeacherLastName); err != nil {
			return nil, err
		}
		return &theme, nil
	}

	return nil, nil

}

func GetTeachers() ([]models.Teacher, error) {
	rows, err := db.Query(`
		SELECT teacher_id, first_name, middle_name, last_name
		FROM teacher
`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var teachers []models.Teacher
	for rows.Next() {
		var teacher models.Teacher
		if err := rows.Scan(&teacher.TeacherID, &teacher.FirstName, &teacher.MiddleName, &teacher.LastName); err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}
	return teachers, nil
}

func SaveTheme(id string, title string, teacher_id string) error {
	query, err := db.Prepare(
		`UPDATE theme
    		SET theme_title = ?, teacher_id = ?
    		WHERE theme_id = ?`)
	if err != nil {
		logging.Error("Ошибка при подготовке запроса: ", err)
		return err
	}
	defer query.Close()

	_, err = query.Exec(title, teacher_id, id)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return err
	}

	return nil
}

func GetThemeToAssign(id string) (*models.Theme, error) {
	rows, err := db.Query(`
		SELECT t.theme_id, t.theme_title, te.first_name, te.middle_name, te.last_name, s.student_id, s.first_name, s.last_name, s.student_record_book
		FROM theme t
		LEFT JOIN teacher te ON t.teacher_id = te.teacher_id
		LEFT JOIN student s ON t.student_record_book = s.student_record_book
		WHERE t.theme_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var theme models.Theme
	if rows.Next() {
		if err := rows.Scan(&theme.ThemeID, &theme.Title, &theme.TeacherFirstName, &theme.TeacherMiddleName, &theme.TeacherLastName, &theme.StudentID, &theme.StudentFirstName, &theme.StudentLastName, &theme.StudentRecordBook); err != nil {
			return nil, err
		}
		return &theme, nil
	}

	return nil, nil
}

func GetStudents() ([]models.Student, error) {
	rows, err := db.Query(`SELECT student_id, first_name, last_name, student_record_book FROM student WHERE (SELECT COUNT(*) FROM theme WHERE student_record_book = student.student_record_book) = 0`)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student
		if err := rows.Scan(&student.StudentID, &student.FirstName, &student.LastName, &student.RecordBook); err != nil {
			logging.Error("Ошибка при сканировании строки: ", err)
			return nil, err
		}
		students = append(students, student)
	}
	return students, nil
}

func AssignTheme(studentRecordBook string, themeID string) error {
	query, err := db.Prepare(
		`UPDATE theme
			SET student_record_book = ?
			WHERE theme_id = ?`)
	if err != nil {
		logging.Error("Ошибка при подготовке запроса: ", err)
		return err
	}

	_, err = query.Exec(studentRecordBook, themeID)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return err
	}

	return nil
}

func GetThemeToGrade(id string) (*models.Theme, error) {
	rows, err := db.Query(`
		SELECT t.theme_id, t.theme_title, te.first_name, te.middle_name, te.last_name, s.student_id, s.first_name, s.last_name, s.student_record_book, m.test_mark, m.diplom_mark
		FROM theme t
		LEFT JOIN teacher te ON t.teacher_id = te.teacher_id
		LEFT JOIN student s ON t.student_record_book = s.student_record_book
		LEFT JOIN mark m ON m.student_record_book = s.student_record_book
		WHERE t.theme_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var theme models.Theme
	if rows.Next() {
		if err := rows.Scan(&theme.ThemeID, &theme.Title, &theme.TeacherFirstName, &theme.TeacherMiddleName, &theme.TeacherLastName, &theme.StudentID, &theme.StudentFirstName, &theme.StudentLastName, &theme.StudentRecordBook, &theme.TestMark, &theme.DiplomMark); err != nil {
			return nil, err
		}
		return &theme, nil
	}

	return nil, nil
}

func SaveGrade(themeID string, testMark int, diplomMark int) error {
	rows, err := db.Query(`
	SELECT test_mark, diplom_mark
	FROM mark
	WHERE student_record_book = (SELECT student_record_book FROM theme WHERE theme_id = ?)`, themeID)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return err
	}

	if rows.Next() {
		_, err := db.Exec(`
		UPDATE mark
		SET test_mark = ?, diplom_mark = ?
		WHERE student_record_book = (SELECT student_record_book FROM theme WHERE theme_id = ?)`, testMark, diplomMark, themeID)
		if err != nil {
			logging.Error("Ошибка при выполнении запроса: ", err)
			return err
		}
	} else {
		_, err := db.Exec(`
		INSERT INTO mark (student_record_book, test_mark, diplom_mark)
		VALUES ((SELECT student_record_book FROM theme WHERE theme_id = ?), ?, ?)`, themeID, testMark, diplomMark)
		if err != nil {
			logging.Error("Ошибка при выполнении запроса: ", err)
			return err
		}
	}
	return nil
}

func AddTheme(title string, id string) error {
	query, err := db.Prepare(
		`INSERT INTO theme (theme_title, teacher_id)
    		VALUES (?, ?)`)
	if err != nil {
		logging.Error("Ошибка при подготовке запроса: ", err)
		return err
	}
	defer query.Close()

	_, err = query.Exec(title, id)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return err
	}
	return nil
}

func GetGrades() ([]models.Grade, error) {
	rows, err := db.Query(`SELECT grade_id, grade_name FROM grade`)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return nil, err
	}
	defer rows.Close()

	var grades []models.Grade
	for rows.Next() {
		var grade models.Grade
		if err := rows.Scan(&grade.GradeID, &grade.GradeName); err != nil {
			logging.Error("Ошибка при сканировании строки: ", err)
			return nil, err
		}
		grades = append(grades, grade)
	}
	return grades, nil
}

func GetAcademicTitles() ([]models.AcademicTitle, error) {
	rows, err := db.Query(`SELECT academic_title_id, academic_title_name FROM academic_title`)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return nil, err
	}
	defer rows.Close()

	var academicTitles []models.AcademicTitle
	for rows.Next() {
		var academicTitle models.AcademicTitle
		if err := rows.Scan(&academicTitle.AcademicTitleID, &academicTitle.AcademicTitleName); err != nil {
			logging.Error("Ошибка при сканировании строки: ", err)
			return nil, err
		}
		academicTitles = append(academicTitles, academicTitle)
	}
	return academicTitles, nil
}

func GetDepartments() ([]models.Department, error) {
	rows, err := db.Query(`SELECT department_id, department_name FROM department`)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return nil, err
	}
	defer rows.Close()

	var departments []models.Department
	for rows.Next() {
		var department models.Department
		if err := rows.Scan(&department.DepartmentID, &department.DepartmentName); err != nil {
			logging.Error("Ошибка при сканировании строки: ", err)
			return nil, err
		}
		departments = append(departments, department)
	}
	return departments, nil
}

func GetTeacherToEdit(id string) (*models.Teacher, error) {
	query, err := db.Query(`
		SELECT teacher_id, first_name, middle_name, last_name, phone, email,  g.grade_id, g.grade_name, at.academic_title_id, at.academic_title_name 
		FROM teacher
		LEFT JOIN grade g ON teacher.grade_id = g.grade_id
		LEFT JOIN academic_title at ON teacher.academic_title_id = at.academic_title_id
		WHERE teacher_id = ?`, id)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return nil, err
	}
	defer query.Close()

	var teacher models.Teacher
	var grade models.Grade
	var academicTitle models.AcademicTitle

	if query.Next() {
		if err := query.Scan(&teacher.TeacherID, &teacher.FirstName, &teacher.MiddleName, &teacher.LastName, &teacher.Phone, &teacher.Email, &grade.GradeID, &grade.GradeName, &academicTitle.AcademicTitleID, &academicTitle.AcademicTitleName); err != nil {
			logging.Error("Ошибка при сканировании строки: ", err)
			return nil, err
		}
		teacher.Grade = &grade
		teacher.AcademicTitle = &academicTitle
	}

	if &teacher != nil {
		query2, err := db.Query(`SELECT department_id, department_name FROM department WHERE department_id IN (SELECT department_id FROM teacher_department WHERE teacher_id = ?)`, id)
		if err != nil {
			logging.Error("Ошибка при выполнении запроса: ", err)
			return nil, err
		}
		defer query2.Close()
		var departments []models.Department
		for query2.Next() {
			var department models.Department
			if err := query2.Scan(&department.DepartmentID, &department.DepartmentName); err != nil {
				logging.Error("Ошибка при сканировании строки: ", err)
				return nil, err
			}
			departments = append(departments, department)
		}
		teacher.Departments = departments
		return &teacher, nil
	}

	return nil, fmt.Errorf("Преподаватель с id %s не найден", id)
}

func SaveTeacher(id string, teacherFirstName string, teacherMiddleName string, teacherLastName string, teacherEmail string, teacherPhone string, gradeID string, academicTitleID string, departmentIDs []string) error {
	query, err := db.Prepare(
		`UPDATE teacher
    		SET first_name = ?, middle_name = ?, last_name = ?, email = ?, phone = ?, grade_id = ?, academic_title_id = ?
    		WHERE teacher_id = ?
    		`)
	if err != nil {
		logging.Error("Ошибка при подготовке запроса: ", err)
		return err
	}
	defer query.Close()

	var academicTitleIDnull sql.NullString
	var gradeIDnull sql.NullString

	if gradeID == "" {
		gradeIDnull = sql.NullString{}
	} else {
		gradeIDnull = sql.NullString{
			String: gradeID,
			Valid:  true,
		}
	}
	if academicTitleID == "" {
		academicTitleIDnull = sql.NullString{}
	} else {
		academicTitleIDnull = sql.NullString{
			String: academicTitleID,
			Valid:  true,
		}
	}

	_, err = query.Exec(teacherFirstName, teacherMiddleName, teacherLastName, teacherEmail, teacherPhone, gradeIDnull, academicTitleIDnull, id)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return err
	}

	logging.Info(fmt.Sprintf("departmentIDs: %v", departmentIDs))

	_, err = db.Exec(`DELETE FROM teacher_department WHERE teacher_id = ?`, id)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return err
	}

	for _, departmentID := range departmentIDs {
		deptID, err := strconv.Atoi(departmentID)
		if err != nil {
			logging.Error("Ошибка при преобразовании строки в число", err)
			return err
		}
		_, err = db.Exec(`INSERT INTO teacher_department (teacher_id, department_id) VALUES (?, ?)`, id, deptID)
		if err != nil {
			logging.Error("Ошибка при выполнении запроса: ", err)
			return err
		}
	}

	return nil

}

func GetThemeByStudent(studentRecordBook string) (*models.Theme, error) {
	rows, err := db.Query(`
		SELECT t.theme_id
		FROM theme t
		WHERE t.student_record_book = ?
		`, studentRecordBook)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var theme models.Theme
	if rows.Next() {
		if err := rows.Scan(&theme.ThemeID); err != nil {
			return nil, err
		}
		return &theme, nil
	}

	return nil, nil

}

func GetGroups() ([]models.Group, error) {
	rows, err := db.Query(fmt.Sprint(`
		SELECT ` + "`group_id`, `group_name`" + `
		FROM` + "`group`" + `
	`))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.GroupID, &group.GroupName); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	return groups, nil
}

func GetFaculties() ([]models.Faculty, error) {
	rows, err := db.Query(`
		SELECT faculty_id, faculty_name
		FROM faculty
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var faculties []models.Faculty
	for rows.Next() {
		var faculty models.Faculty
		if err := rows.Scan(&faculty.FacultyID, &faculty.FacultyName); err != nil {
			return nil, err
		}
		faculties = append(faculties, faculty)
	}

	return faculties, nil
}

func GetStudentToEdit(id string) (*models.Student, error) {
	rows, err := db.Query(fmt.Sprint(`
    SELECT s.student_id, s.first_name, s.last_name, s.student_record_book, s.faculty_id,`+"`group_id`"+`
    FROM student s
    WHERE s.student_id = ?
`), id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var student models.Student
	if rows.Next() {
		if err := rows.Scan(&student.StudentID, &student.FirstName, &student.LastName, &student.RecordBook, &student.FacultyID, &student.GroupID); err != nil {
			return nil, err
		}
		return &student, nil
	}

	return nil, nil
}

func DeleteTeacher(id string) error {
	_, err := db.Exec(`DELETE FROM teacher WHERE teacher_id = ?`, id)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return err
	}
	return nil
}

func SaveStudent(id string, firstName string, lastName string, recordBook string, facultyID string, groupID string) error {
	_, err := db.Exec(`
	UPDATE student
	SET first_name = ?, last_name = ?, student_record_book = ?, faculty_id = ?, `+"`group_id`"+` = ?
	WHERE student_id = ?`, firstName, lastName, recordBook, facultyID, groupID, id)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return err
	}

	return nil
}

func DeleteStudent(id string) error {
	_, err := db.Exec(`DELETE FROM student WHERE student_id = ?`, id)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return err
	}
	return nil
}

func DeleteTheme(id string) error {
	_, err := db.Exec(`DELETE FROM theme WHERE theme_id = ?`, id)
	if err != nil {
		logging.Error("Ошибка при выполнении запроса: ", err)
		return err
	}
	return nil
}

func GetGroupsByFaculty(facultyID int) ([]models.Group, error) {
	rows, err := db.Query(`
		SELECT DISTINCT g.group_id, g.group_name
		FROM student s
		JOIN `+"`group`"+` g ON s.group_id = g.group_id
		WHERE s.faculty_id = ?`, facultyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.GroupID, &group.GroupName); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func GetStudentsByGroup(groupID int) ([]models.Student, error) {
	rows, err := db.Query(`
		SELECT s.student_record_book, s.first_name, s.last_name
		FROM student s
		WHERE s.group_id = ?`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student
		if err := rows.Scan(&student.RecordBook, &student.FirstName, &student.LastName); err != nil {
			return nil, err
		}
		students = append(students, student)
	}
	return students, nil
}

func GetMarksByStudent(studentRecordBook string) (sql.NullInt64, sql.NullInt64, error) {
	row := db.QueryRow(`
		SELECT test_mark, diplom_mark 
		FROM mark 
		WHERE student_record_book = ?`, studentRecordBook)

	var testMark, diplomMark sql.NullInt64
	err := row.Scan(&testMark, &diplomMark)
	if err == sql.ErrNoRows {
		return sql.NullInt64{}, sql.NullInt64{}, nil
	} else if err != nil {
		return sql.NullInt64{}, sql.NullInt64{}, err
	}
	return testMark, diplomMark, nil
}
