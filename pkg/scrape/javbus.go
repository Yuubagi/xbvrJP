package scrape

import (
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/xbapps/xbvr/pkg/models"
)

func ScrapeJavBus(out *[]models.ScrapedScene, queryString string) {
	sceneCollector := createCollector("www.javbus.com")

	sceneCollector.OnHTML(`html`, func(html *colly.HTMLElement) {
		sc := models.ScrapedScene{}
		sc.SceneType = "VR"

		html.ForEach(`div.row.movie div.info > p`, func(id int, p *colly.HTMLElement) {
			label := p.ChildText(`span.header`)

			if label == `メーカー:` {
				// Studio
				sc.Studio = p.ChildText(`a`)

			} else if label == `品番:` {
				// Title, SceneID and SiteID all like 'VRKM-821' format
				idRegex := regexp.MustCompile("^([A-Za-z0-9]+)-([0-9]+)$")
				p.ForEach("span", func(_ int, span *colly.HTMLElement) {
					match := idRegex.FindStringSubmatch(span.Text)
					if match != nil && len(match) > 2 {
						dvdId := match[1] + "-" + match[2]
						sc.SceneID = dvdId
						sc.SiteID = dvdId
						sc.Site = match[1]

					}
				})

			} else if label == `発売日:` {
				// Release date
				dateStr := p.Text
				dateRegex := regexp.MustCompile("(\\d\\d\\d\\d-\\d\\d-\\d\\d)")
				match := dateRegex.FindStringSubmatch(dateStr)
				if match != nil && len(match) > 1 {
					sc.Released = match[1]
				}
			}
		})
		//Japn Title
		html.ForEach(`div.container > h3`, func(_ int, elem *colly.HTMLElement) {
    			jpntitle := elem.Text
			titleWithoutDvdId := strings.ReplaceAll(jpntitle, sc.SceneID + " ", "")
			sc.Title = titleWithoutDvdId
		})
		// Genres
		html.ForEach("div.row.movie span.genre > label > a", func(id int, anchor *colly.HTMLElement) {
   			href := anchor.Attr("href")
    			if strings.Contains(href, "javbus.com/ja/genre/") {
        			// Genres
        			genre := strings.TrimSpace(anchor.Text)

        			if genre != "" {
            			sc.Tags = append(sc.Tags, genre)
        			}
    			}
		})
		// Cast
		html.ForEach("div.row.movie div.star-name > a", func(id int, anchor *colly.HTMLElement) {
			href := anchor.Attr("href")
			if strings.Contains(href, "javbus.com/ja/star/") {
				sc.Cast = append(sc.Cast, anchor.Text)
			}
		})

		// Screenshots
		html.ForEach("img[src]", func(_ int, img *colly.HTMLElement) {
			src := img.Attr("src")
			if strings.HasPrefix(src, "https://pics.dmm.co.jp/digital/video/") && strings.HasSuffix(src, ".jpg") {
				sc.Gallery = append(sc.Gallery, src)
			}
		})

		// Apply post-processing for error-correcting code
		PostProcessJavScene(&sc, "")

		if sc.SceneID != "" {
			*out = append(*out, sc)
		}
	})

	// Allow comma-separated scene id's
	scenes := strings.Split(queryString, ",")
	for _, v := range scenes {
		sceneCollector.Visit("https://www.javbus.com/ja/" + strings.ToUpper(v) + "/")
	}

	sceneCollector.Wait()
}
