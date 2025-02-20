package main

import (
	"fmt"
	"github.com/PIRSON21/lab3/db"
	"github.com/PIRSON21/lab3/logging"
	"github.com/PIRSON21/lab3/models"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type Error struct {
	Message string
}

var error Error

type FacultyData struct {
	FacultyName string
	Groups      []GroupData
	FacultyAvg  float64
}

type GroupData struct {
	GroupName string
	Students  []StudentData
	GroupAvg  float64
}

type StudentData struct {
	StudentFullName string
	TestMark        int
	DiplomMark      int
	AverageMark     float64
}

// Главная страница: выводим таблицу с темами
func themesHandler(w http.ResponseWriter, r *http.Request) {
	// Логируем запрос
	logging.Info("Запрос на страницу с темами")

	filter := r.URL.Query().Get("filter")

	// Получаем темы из базы данных
	themes, err := db.GetThemes(filter)
	if err != nil {
		logging.Error("Ошибка при получении данных о темах: ", err)
		error.Message = "Ошибка при получении данных"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := struct {
		Themes []models.Theme
		Error  Error
		Filter string
	}{themes, error, filter}

	// Загружаем шаблон и передаем данные в шаблон
	tmpl, err := template.ParseFiles("templates/index.tmpl")
	if err != nil {
		logging.Error("Ошибка при загрузке шаблона: ", err)
		http.Error(w, "Ошибка при загрузке шаблона", http.StatusInternalServerError)
		return
	}

	// Отправляем шаблон в ответ
	err = tmpl.Execute(w, data)
	if err != nil {
		logging.Error("Ошибка при рендеринге шаблона: ", err)
		http.Error(w, "Ошибка при рендеринге страницы", http.StatusInternalServerError)
	}
	error.Message = ""
}

func themeEditHandler(w http.ResponseWriter, r *http.Request) {
	// Логируем запрос
	logging.Info("Запрос на страницу редактирования темы")

	// Получаем параметр из URL
	id := r.URL.Path[len("/editTheme/"):]
	logging.Info(fmt.Sprint("ID темы: ", id))

	// Получаем тему из базы данных
	theme, err := db.GetThemeToEdit(id)
	if err != nil {
		logging.Error("Ошибка при получении данных о теме: ", err)
		http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
		return
	}

	// Получаем список преподавателей из базы данных
	teachers, err := db.GetTeachers()
	if err != nil {
		logging.Error("Ошибка при получении данных о преподавателях: ", err)
		http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
		return
	}

	data := struct {
		Theme    *models.Theme
		Teachers []models.Teacher
		Error    Error
	}{theme, teachers, error}

	// Загружаем шаблон и передаем данные в шаблон
	tmpl, err := template.ParseFiles("templates/theme_edit.tmpl")
	if err != nil {
		logging.Error("Ошибка при загрузке шаблона: ", err)
		http.Error(w, "Ошибка при загрузке шаблона", http.StatusInternalServerError)
		return
	}

	// Отправляем шаблон в ответ
	err = tmpl.Execute(w, data)
	if err != nil {
		logging.Error("Ошибка при рендеринге шаблона: ", err)
		http.Error(w, "Ошибка при рендеринге страницы", http.StatusInternalServerError)
	}
	error.Message = ""
}

func saveThemeHandler(w http.ResponseWriter, r *http.Request) {
	// Логируем запрос
	logging.Info("Запрос на сохранение темы")

	if r.Method == http.MethodPost {
		r.ParseForm()
		themeID := r.PostFormValue("theme_id")
		themeName := r.PostFormValue("theme_name")
		teacherID := r.PostFormValue("teacher_id")

		prevTheme, err := db.GetThemeToAssign(themeID)
		if err != nil {
			logging.Error("Ошибка при получении данных о теме: ", err)
			error.Message = "Ошибка при получении данных"
			http.Redirect(w, r, "/editTheme/"+themeID, http.StatusSeeOther)
			return
		}
		if prevTheme.DiplomMark.Valid || prevTheme.TestMark.Valid {
			error.Message = "Нельзя изменить тему, которая уже оценена"
			http.Redirect(w, r, "/editTheme/"+themeID, http.StatusSeeOther)
			return
		}
		if prevTheme.StudentRecordBook.Valid {
			error.Message = "Нельзя изменить тему, которая уже назначена студенту"
			http.Redirect(w, r, "/editTheme/"+themeID, http.StatusSeeOther)
			return
		}

		teacher, err := db.GetTeacherToEdit(teacherID)
		if err != nil {
			logging.Error("Ошибка при получении данных о преподавателе: ", err)
			error.Message = "Ошибка при получении данных"
			http.Redirect(w, r, "/editTheme/"+themeID, http.StatusSeeOther)
			return
		}

		if !teacher.AcademicTitle.AcademicTitleID.Valid {
			error.Message = "Преподаватель не имеет ученого звания"
			http.Redirect(w, r, "/editTheme/"+themeID, http.StatusSeeOther)
			return
		}

		// Сохраняем тему в базе данных
		err = db.SaveTheme(themeID, themeName, teacherID)
		if err != nil {
			logging.Error("Ошибка при сохранении данных о теме: ", err)
			error.Message = "Ошибка при сохранении данных"
			http.Redirect(w, r, "/editTheme/"+themeID, http.StatusSeeOther)
		}

		// Перенаправляем на главную страницу
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func assignThemeHandler(w http.ResponseWriter, r *http.Request) {
	// Логируем запрос
	logging.Info("Запрос на назначение темы")

	if r.Method == http.MethodPost {
		r.ParseForm()
		studentRecordBook := r.PostFormValue("student_record_book")
		themeID := r.PostFormValue("theme_id")

		cur_theme, err := db.GetThemeToGrade(themeID)
		if err != nil {
			logging.Error("Ошибка при получении данных о теме: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}
		if cur_theme.StudentRecordBook.Valid && cur_theme.StudentRecordBook.String != studentRecordBook && (cur_theme.DiplomMark.Valid || cur_theme.TestMark.Valid) {
			error.Message = "Нельзя изменить тему, которая уже оценена"
			http.Redirect(w, r, "/assignTheme/"+themeID, http.StatusSeeOther)
			return
		}

		students_theme, err := db.GetThemeByStudent(studentRecordBook)
		if err != nil {
			logging.Error("Ошибка при получении данных о теме: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}
		themeIDint, err := strconv.Atoi(themeID)
		if err != nil {
			logging.Error("Ошибка при преобразовании ID темы: ", err)
			http.Error(w, "Ошибка при преобразовании данных", http.StatusInternalServerError)
			return
		}
		if students_theme != nil && students_theme.ThemeID != themeIDint {
			error.Message = "Студент уже назначен на другую тему"
			http.Redirect(w, r, "/assignTheme/"+themeID, http.StatusSeeOther)
			return
		}

		// Назначаем тему студенту в базе данных
		err = db.AssignTheme(studentRecordBook, themeID)
		if err != nil {
			logging.Error("Ошибка при назначении темы: ", err)
			error.Message = "Ошибка при назначении темы"
			http.Redirect(w, r, "/assignTheme/"+themeID, http.StatusSeeOther)
			return
		}

		// Перенаправляем на главную страницу
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else if r.Method == http.MethodGet {
		// Получаем параметр из URL
		id := r.URL.Path[len("/assignTheme/"):]
		logging.Info(fmt.Sprint("ID темы: ", id))

		// Получаем тему из базы данных
		theme, err := db.GetThemeToAssign(id)
		if err != nil {
			logging.Error("Ошибка при получении данных о теме: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		// Получаем список студентов из базы данных
		students, err := db.GetStudents()
		if err != nil {
			logging.Error("Ошибка при получении данных о студентах: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		data := struct {
			Theme    *models.Theme
			Students []models.Student
			Error    Error
		}{theme, students, error}

		// Загружаем шаблон и передаем данные в шаблон
		tmpl, err := template.ParseFiles("templates/assign_theme.tmpl")
		if err != nil {
			logging.Error("Ошибка при загрузке шаблона: ", err)
			http.Error(w, "Ошибка при загрузке шаблона", http.StatusInternalServerError)
			return
		}

		// Отправляем шаблон в ответ
		err = tmpl.Execute(w, data)
		if err != nil {
			logging.Error("Ошибка при рендеринге шаблона: ", err)
			http.Error(w, "Ошибка при рендеринге страницы", http.StatusInternalServerError)
		}
		error.Message = ""
	}
}

func editGradeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Логируем запрос
		logging.Info("Запрос на страницу редактирования оценок")

		// Получаем параметр из URL
		id := r.URL.Path[len("/grade/"):]
		logging.Info(fmt.Sprint("ID темы: ", id))

		// Получаем тему из базы данных
		theme, err := db.GetThemeToGrade(id)
		if err != nil {
			logging.Error("Ошибка при получении данных о теме: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		if !theme.StudentRecordBook.Valid {
			error.Message = "Тема не назначена студенту"
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		data := struct {
			Theme *models.Theme
			Error Error
		}{theme, error}

		// Загружаем шаблон и передаем данные в шаблон
		tmpl, err := template.ParseFiles("templates/edit_grade.tmpl")
		if err != nil {
			logging.Error("Ошибка при загрузке шаблона: ", err)
			http.Error(w, "Ошибка при загрузке шаблона", http.StatusInternalServerError)
			return
		}

		// Отправляем шаблон в ответ
		err = tmpl.Execute(w, data)
		if err != nil {
			logging.Error("Ошибка при рендеринге шаблона: ", err)
			http.Error(w, "Ошибка при рендеринге страницы", http.StatusInternalServerError)
			return
		}
		error.Message = ""
	} else if r.Method == http.MethodPost {
		// Логируем запрос
		logging.Info("Запрос на сохранение оценок")

		r.ParseForm()
		themeID := r.PostFormValue("theme_id")
		tm := r.PostFormValue("test_mark")
		dm := r.PostFormValue("diplom_mark")

		// Преобразуем оценки в числовой формат
		testMark, err := strconv.Atoi(tm)
		if err != nil {
			logging.Error("Ошибка при преобразовании оценки за тест: ", err)
			http.Error(w, "Ошибка при преобразовании данных", http.StatusInternalServerError)
			return
		}
		diplomMark, err := strconv.Atoi(dm)
		if err != nil {
			logging.Error("Ошибка при преобразовании оценки за диплом: ", err)
			http.Error(w, "Ошибка при преобразовании данных", http.StatusInternalServerError)
			return
		}

		if testMark < 2 || testMark > 5 {
			logging.Info(fmt.Sprint("Оценка за тестовую работу не действительна", testMark))
			error.Message = "Оценка за тестовую работу не действительна"
			http.Redirect(w, r, "/grade/"+themeID, http.StatusSeeOther)
			return
		}

		if diplomMark < 2 || diplomMark > 5 {
			logging.Info(fmt.Sprint("Оценка за тестовую работу не действительна", testMark))
			error.Message = "Оценка за дипломную работу не действительна"
			http.Redirect(w, r, "/grade/"+themeID, http.StatusSeeOther)
			return
		}

		theme, err := db.GetThemeToGrade(themeID)
		if err != nil {
			logging.Error("Ошибка при получении данных о теме: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}
		if !theme.StudentRecordBook.Valid {
			error.Message = "Тема не назначена студенту"
			http.Redirect(w, r, "/grade/"+themeID, http.StatusSeeOther)
		}

		// Сохраняем оценки в базе данных
		err = db.SaveGrade(themeID, testMark, diplomMark)
		if err != nil {
			logging.Error("Ошибка при сохранении данных оценок: ", err)
			error.Message = "Ошибка при сохранении данных"
			http.Redirect(w, r, "/grade/"+themeID, http.StatusSeeOther)
			return
		}

		// Перенаправляем на главную страницу
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func addThemeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet { // Логируем запрос
		logging.Info("Запрос на страницу добавления темы")

		// Получаем список преподавателей из базы данных
		teachers, err := db.GetTeachers()
		if err != nil {
			logging.Error("Ошибка при получении данных о преподавателях: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		data := struct {
			Teachers []models.Teacher
			Error    Error
		}{teachers, error}

		// Загружаем шаблон и передаем данные в шаблон
		tmpl, err := template.ParseFiles("templates/add_theme.tmpl")
		if err != nil {
			logging.Error("Ошибка при загрузке шаблона: ", err)
			http.Error(w, "Ошибка при загрузке шаблона", http.StatusInternalServerError)
			return
		}

		// Отправляем шаблон в ответ
		err = tmpl.Execute(w, data)
		if err != nil {
			logging.Error("Ошибка при рендеринге шаблона: ", err)
			http.Error(w, "Ошибка при рендеринге страницы", http.StatusInternalServerError)
		}
	} else if r.Method == http.MethodPost {
		// Логируем запрос
		logging.Info("Запрос на добавление темы")

		r.ParseForm()
		themeTitle := r.PostFormValue("theme_title")
		teacherID := r.PostFormValue("teacher_id")

		teacher, err := db.GetTeacherToEdit(teacherID)
		if err != nil {
			logging.Error("Ошибка при получении данных о преподавателе: ", err)
			error.Message = "Ошибка при получении данных"
			http.Redirect(w, r, "/addTheme/", http.StatusSeeOther)
			return
		}

		if !teacher.AcademicTitle.AcademicTitleID.Valid {
			error.Message = "Преподаватель не имеет ученого звания"
			http.Redirect(w, r, "/addTheme/", http.StatusSeeOther)
			return
		}

		// Добавляем тему в базу данных
		err = db.AddTheme(themeTitle, teacherID)
		if err != nil {
			logging.Error("Ошибка при добавлении темы: ", err)
			error.Message = "Ошибка при добавлении темы"
			http.Redirect(w, r, "/addTheme/", http.StatusSeeOther)
			return
		}

		// Перенаправляем на главную страницу
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func editTeacherHandler(w http.ResponseWriter, r *http.Request) {
	// Логируем запрос
	logging.Info("Запрос на страницу редактирования преподавателя")

	if r.Method == http.MethodGet {
		// Получаем параметр из URL
		id := r.URL.Path[len("/teacher/"):]
		logging.Info(fmt.Sprint("ID преподавателя: ", id))

		// Получаем преподавателя из базы данных
		teacher, err := db.GetTeacherToEdit(id)
		if err != nil {
			logging.Error("Ошибка при получении данных о преподавателе: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		grades, err := db.GetGrades()
		if err != nil {
			logging.Error("Ошибка при получении данных званий: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		academicTitles, err := db.GetAcademicTitles()
		if err != nil {
			logging.Error("Ошибка при получении данных степеней: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		departments, err := db.GetDepartments()
		if err != nil {
			logging.Error("Ошибка при получении данных кафедр: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		selectedDept := make(map[int]bool)
		for _, dept := range teacher.Departments {
			selectedDept[dept.DepartmentID] = true
		}

		data := struct {
			Teacher        *models.Teacher
			Grades         []models.Grade
			AcademicTitles []models.AcademicTitle
			Departments    []models.Department
			SelectedDept   map[int]bool
			Error          Error
		}{teacher, grades, academicTitles, departments, selectedDept, error}

		logging.Info(fmt.Sprint("Департаменты", departments))
		// Загружаем шаблон и передаем данные в шаблон
		tmpl, err := template.ParseFiles("templates/teacher_edit.tmpl")
		if err != nil {
			logging.Error("Ошибка при загрузке шаблона: ", err)
			http.Error(w, "Ошибка при загрузке шаблона", http.StatusInternalServerError)
			return
		}

		// Отправляем шаблон в ответ
		err = tmpl.Execute(w, data)
		if err != nil {
			logging.Error("Ошибка при рендеринге шаблона: ", err)
			http.Error(w, "Ошибка при рендеринге страницы", http.StatusInternalServerError)
		}
		error.Message = ""
	} else if r.Method == http.MethodPost {
		// Логируем запрос
		logging.Info("Запрос на сохранение данных преподавателя")

		r.ParseForm()
		teacherID := r.PostFormValue("teacher_id")
		teacherFirstName := r.PostFormValue("first_name")
		teacherMiddleName := r.PostFormValue("middle_name")
		teacherLastName := r.PostFormValue("last_name")
		teacherPhone := r.PostFormValue("phone")
		teacherEmail := r.PostFormValue("email")
		gradeID := r.PostFormValue("grade_id")
		academicTitleID := r.PostFormValue("academic_title_id")
		departmentIDs := r.PostForm["department_ids"]

		// Сохраняем данные преподавателя в базе данных
		err := db.SaveTeacher(teacherID, teacherFirstName, teacherMiddleName, teacherLastName, teacherEmail, teacherPhone, gradeID, academicTitleID, departmentIDs)
		if err != nil {
			logging.Error("Ошибка при сохранении данных преподавателя: ", err)
			http.Error(w, "Ошибка при сохранении данных", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на главную страницу
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func deleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	// Логируем запрос
	logging.Info("Запрос на удаление преподавателя")

	// Получаем параметр из URL
	id := r.URL.Path[len("/delete-teacher/"):]
	logging.Info(fmt.Sprint("ID преподавателя: ", id))

	// Удаляем преподавателя из базы данных
	err := db.DeleteTeacher(id)
	if err != nil {
		logging.Error("Ошибка при удалении преподавателя: ", err)
		http.Error(w, "Ошибка при удалении преподавателя", http.StatusInternalServerError)
		return
	}

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editStudentHandler(w http.ResponseWriter, r *http.Request) {
	// Логируем запрос
	logging.Info("Запрос на страницу редактирования студента")

	if r.Method == http.MethodGet {
		// Получаем параметр из URL
		id := r.URL.Path[len("/student/"):]
		logging.Info(fmt.Sprint("ID студента: ", id))

		// Получаем студента из базы данных
		student, err := db.GetStudentToEdit(id)
		if err != nil {
			logging.Error("Ошибка при получении данных о студенте: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		logging.Info(fmt.Sprint("Студент: ", student))

		groups, err := db.GetGroups()
		if err != nil {
			logging.Error("Ошибка при получении данных групп: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		logging.Info(fmt.Sprint("Группы: ", groups))

		faculties, err := db.GetFaculties()
		if err != nil {
			logging.Error("Ошибка при получении данных факультетов: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		logging.Info(fmt.Sprint("Факультеты: ", faculties))

		data := struct {
			Student   *models.Student
			Groups    []models.Group
			Faculties []models.Faculty
			Error     Error
		}{student, groups, faculties, error}

		// Загружаем шаблон и передаем данные в шаблон
		tmpl, err := template.ParseFiles("templates/student_edit.tmpl")
		if err != nil {
			logging.Error("Ошибка при загрузке шаблона: ", err)
			http.Error(w, "Ошибка при загрузке шаблона", http.StatusInternalServerError)
			return
		}

		// Отправляем шаблон в ответ
		err = tmpl.Execute(w, data)
		if err != nil {
			logging.Error("Ошибка при рендеринге шаблона: ", err)
			http.Error(w, "Ошибка при рендеринге страницы", http.StatusInternalServerError)
		}
		error.Message = ""
	} else if r.Method == http.MethodPost {
		logging.Info("Запрос на сохранение данных студента")

		r.ParseForm()
		studentID := r.PostFormValue("student_id")
		studentFirstName := r.PostFormValue("first_name")
		studentLastName := r.PostFormValue("last_name")
		studentRecordBook := r.PostFormValue("record_book")
		facultyID := r.PostFormValue("faculty_id")
		groupID := r.PostFormValue("group_id")

		err := db.SaveStudent(studentID, studentFirstName, studentLastName, studentRecordBook, facultyID, groupID)
		if err != nil {
			logging.Error("Ошибка при сохранении данных студента: ", err)
			http.Error(w, "Ошибка при сохранении данных", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

	}
}

func deleteStudentHandler(w http.ResponseWriter, r *http.Request) {
	// Логируем запрос
	logging.Info("Запрос на удаление студента")

	// Получаем параметр из URL
	id := r.URL.Path[len("/delete-student/"):]
	logging.Info(fmt.Sprint("ID студента: ", id))

	// Удаляем студента из базы данных
	err := db.DeleteStudent(id)
	if err != nil {
		logging.Error("Ошибка при удалении студента: ", err)
		http.Error(w, "Ошибка при удалении студента", http.StatusInternalServerError)
		return
	}

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteThemeHandler(w http.ResponseWriter, r *http.Request) {
	// Логируем запрос
	logging.Info("Запрос на удаление темы")

	// Получаем параметр из URL
	id := r.URL.Path[len("/delete-theme/"):]
	logging.Info(fmt.Sprint("ID темы: ", id))

	// Удаляем тему из базы данных
	err := db.DeleteTheme(id)
	if err != nil {
		logging.Error("Ошибка при удалении темы: ", err)
		error.Message = "Ошибка при удалении темы"
		http.Redirect(w, r, "/editTheme/"+id, http.StatusSeeOther)
		return
	}

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

//goland:noinspection t
func getReportHandler(w http.ResponseWriter, r *http.Request) {
	faculties, err := db.GetFaculties()
	if err != nil {
		logging.Error("Ошибка при получении данных факультетов: ", err)
		http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
		return
	}

	// Собираем данные для каждого факультета
	var pageData []FacultyData

	for _, faculty := range faculties {
		groups, err := db.GetGroupsByFaculty(faculty.FacultyID)
		if err != nil {
			logging.Error("Ошибка при получении данных групп: ", err)
			http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
			return
		}

		var groupData []GroupData
		var totalFacultyAverage float64
		groupsCount := 0

		for _, group := range groups {
			students, err := db.GetStudentsByGroup(group.GroupID)
			if err != nil {
				logging.Error("Ошибка при получении данных студентов: ", err)
				http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
				return
			}

			var studentData []StudentData
			var groupAverage float64

			for _, student := range students {
				testMark, diplomMark, err := db.GetMarksByStudent(student.RecordBook)
				if err != nil {
					logging.Error("Ошибка при получении данных оценок: ", err)
					http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
					return
				}

				averageMark := (float64(testMark.Int64) + float64(diplomMark.Int64)) / 2
				studentData = append(studentData, StudentData{
					StudentFullName: student.FirstName + " " + student.LastName,
					TestMark:        int(testMark.Int64),
					DiplomMark:      int(diplomMark.Int64),
					AverageMark:     averageMark,
				})

				groupAverage += averageMark
			}

			groupsCount++
			groupAverage /= float64(len(students))
			totalFacultyAverage += groupAverage

			groupData = append(groupData, GroupData{
				GroupName: group.GroupName,
				Students:  studentData,
				GroupAvg:  groupAverage,
			})

		}

		facultyAverage := totalFacultyAverage / float64(groupsCount)

		pageData = append(pageData, FacultyData{
			FacultyName: faculty.FacultyName,
			Groups:      groupData,
			FacultyAvg:  facultyAverage,
		})
	}

	tmpl, err := template.ParseFiles("templates/report.tmpl")
	if err != nil {
		logging.Error("Ошибка при загрузке шаблона: ", err)
		http.Error(w, "Ошибка при загрузке шаблона", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, pageData)
	if err != nil {
		logging.Error("Ошибка при рендеринге шаблона: ", err)
		http.Error(w, "Ошибка при рендеринге страницы", http.StatusInternalServerError)
	}
}

func main() {
	// Инициализируем логирование
	logging.Init()

	// Инициализируем подключение к базе данных
	err := db.InitDB()
	if err != nil {
		logging.Fatal("Не удалось подключиться к базе данных: ", err)
		return
	}
	defer db.Close()

	// Настроим маршруты
	http.HandleFunc("/editTheme/", themeEditHandler)
	http.HandleFunc("/saveTheme/", saveThemeHandler)
	http.HandleFunc("/assignTheme/", assignThemeHandler)
	http.HandleFunc("/grade/", editGradeHandler)
	http.HandleFunc("/addTheme/", addThemeHandler)
	http.HandleFunc("/teacher/", editTeacherHandler)
	http.HandleFunc("/student/", editStudentHandler)
	http.HandleFunc("/delete-teacher/", deleteTeacherHandler)
	http.HandleFunc("/delete-student/", deleteStudentHandler)
	http.HandleFunc("/delete-theme/", deleteThemeHandler)
	http.HandleFunc("/", themesHandler)
	http.HandleFunc("/report/", getReportHandler)

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex) + "\\styles"
	logging.Info(fmt.Sprint("exPath = ", exPath))
	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir(exPath))))

	// Запуск сервера
	logging.Info("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
