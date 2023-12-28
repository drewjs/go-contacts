package main

import (
	"html/template"
	"io"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

var id = 0

type Contact struct {
    Id    int
	Name  string
	Email string
}

func newContact(name string, email string) Contact {
    id++
	return Contact{
        Id:    id,
		Name:  name,
		Email: email,
	}
}

type Contacts = []Contact

type Data struct {
	Contacts Contacts
}

func newData() Data {
	return Data{
		Contacts: Contacts{
			newContact("John", "jd@gmail.com"),
			newContact("Claire", "cd@gmail.com"),
		},
	}
}

func (d *Data) hasEmail(email string) bool {
	for _, c := range d.Contacts {
		if c.Email == email {
			return true
		}
	}
	return false
}

func (d *Data) indexOf(id int) int {
    for i, c := range d.Contacts {
        if c.Id == id {
            return i
        }
    }
    return -1
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	Data Data
	Form FormData
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

    e.Static("/images", "images")
    e.Static("/css", "css")

	e.Renderer = newTemplate()
	page := newPage()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.hasEmail(email) {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists"
			return c.Render(422, "form", formData)
		}

		contact := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contact)

		err := c.Render(200, "contact-form", newFormData())
		if err != nil {
			return err
		}

		return c.Render(200, "oob-contact", contact)
	})

    e.DELETE("/contacts/:id", func(c echo.Context) error {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            return c.String(400, "Invalid id")
        }

        index := page.Data.indexOf(id)
        if index == -1 {
            return c.String(404, "Contact not found")
        }

        page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)

        return c.NoContent(200)
    })

	e.Logger.Fatal(e.Start(":42069"))
}
