package docs

import (
	"html/template"
	"path"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/utils"
	swgFiles "github.com/swaggo/files"
	"github.com/swaggo/swag"
)

const (
	DefaultDocURL = "document.json"
	defaultIndex  = "index.html"
)

type Config struct {
	DeepLinking bool
	URL         string
}

func New(config ...Config) fiber.Handler {
	cfg := Config{
		DeepLinking: true,
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	index, err := template.New("swagger_index.html").Parse(indexTmpl)
	if err != nil {
		panic("swagger: could not parse index template")
	}

	var (
		prefix string
		once   sync.Once
		fs     = filesystem.New(filesystem.Config{Root: swgFiles.HTTP})
	)

	return func(c *fiber.Ctx) error {
		// Set prefix
		once.Do(func() {
			prefix = strings.ReplaceAll(c.Route().Path, "*", "")
			// Set doc url
			cfg.URL = path.Join(prefix, DefaultDocURL)
		})

		var p string
		if p = utils.CopyString(c.Params("*")); p != "" {
			c.Path(p)
		} else {
			p = strings.TrimPrefix(c.Path(), prefix)
			p = strings.TrimPrefix(p, "/")
		}

		switch p {
		case defaultIndex:
			c.Type("html")
			return index.Execute(c, cfg)
		case DefaultDocURL:
			doc, err := swag.ReadDoc()
			if err != nil {
				return err
			}
			return c.Type("json").SendString(doc)
		case "", "/":
			return c.Redirect(path.Join(prefix, defaultIndex), fiber.StatusMovedPermanently)
		default:
			return fs(c)
		}
	}
}
