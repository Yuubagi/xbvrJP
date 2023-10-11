package scrape

import (
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/nleeper/goment"
	"github.com/xbapps/xbvr/pkg/models"
)

func ScrapeJavLand(out *[]models.ScrapedScene, queryString string) {
	sceneCollector := createCollector("jav.land")

	sceneCollector.OnHTML(`html`, func(html *colly.HTMLElement) {
		sc := models.ScrapedScene{}
		sc.SceneType = "VR"
		contentId := ""

		html.ForEach(`table.videotextlist tr`, func(id int, tr *colly.HTMLElement) {
			tds := tr.DOM.Children()
			if tds.Length() != 2 {
				return
			}
			label := tds.First().Text()
			value := tds.Last().Text()

			if label == `メーカー:` {
				// Studio
				sc.Studio = value

			} else if label == `DVD ID:` {
				// Title, SceneID and SiteID all like 'VRKM-821' format
				dvdId := strings.ToUpper(value)
				sc.Synopsis = dvdId
				sc.SceneID = dvdId
				sc.SiteID = dvdId

				// Set 'Site' to first part of the ID (e.g. `VRKM for `vrkm-821`)
				siteParts := strings.Split(dvdId, `-`)
				if len(siteParts) > 0 {
					sc.Site = siteParts[0]
				}

			} else if label == `発売日:` {
				// Release date
				tmpDate, _ := goment.New(strings.TrimSpace(value), "YYYY-MM-DD")
				sc.Released = tmpDate.Format("YYYY-MM-DD")

			} else if label == `ジャンル:` {
				// Tags
				tr.ForEach("span.genre > a", func(id int, anchor *colly.HTMLElement) {
					href := anchor.Attr("href")
    					if strings.Contains(href, "/genre/") {
						sc.Tags = append(sc.Tags, anchor.Text)
    						}
					})

			} else if label == `出演者:` {
				// Tags
				tr.ForEach("span.star > a", func(id int, anchor *colly.HTMLElement) {
					href := anchor.Attr("href")
					if strings.Contains(href, "/star/") {
						sc.Cast = append(sc.Cast, anchor.Text)
					}
				})

			} else if label == `品番:` {
				contentId = value
			}
		})

		// Screenshots
		html.ForEach("a[href]", func(_ int, anchor *colly.HTMLElement) {
			linkHref := anchor.Attr(`href`)
			if strings.HasPrefix(linkHref, "https://pics.vpdmm.cc/") && strings.HasSuffix(linkHref, `.jpg`) {
				linkHref = strings.Replace(linkHref, "https://pics.vpdmm.cc/", "https://pics.dmm.co.jp/", 1)
			}
			if strings.HasPrefix(linkHref, "https://pics.dmm.co.jp/digital/video/") && strings.HasSuffix(linkHref, `.jpg`) {
				sc.Gallery = append(sc.Gallery, linkHref)
			}
		})

		// Synopsis
		title := html.DOM.Find("title")
		if title != nil && title.Length() == 1 {
			descr := title.Text()
			descr = strings.ReplaceAll(descr, "- JAV.Land", "")
			titleWithoutDvdId := strings.ReplaceAll(descr, sc.SceneID + " ", "")
			titleWithoutVR := strings.ReplaceAll(titleWithoutDvdId, "【VR】", "")
			sc.Title = titleWithoutVR
			
		}

		// Apply post-processing for error-correcting code
		PostProcessJavScene(&sc, contentId)

		if sc.SceneID != "" {
			*out = append(*out, sc)
		}
	})

	// Allow comma-separated scene id's
	scenes := strings.Split(queryString, ",")
	for _, v := range scenes {
		sceneCollector.Visit("https://jav.land/ja/id_search.php?keys=" + strings.ToLower(v))
	}

	sceneCollector.Wait()
}
