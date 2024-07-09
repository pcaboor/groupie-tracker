package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Page struct {
	Title string
	Error string

	Data1 []Artist
	Data2 DatesLocation

	Data5           Artist
	Data6           map[string][]string
	ArtistLocations []string

	Filters Filters
}

type Filters struct {
	CreationDateFrom int
	CreationDateTo   int
	FirstAlbumFrom   string
	FirstAlbumTo     string
	MembersFrom      int
	MembersTo        int
	Locations        []string
}

var VarArtists []Artist
var VarDateLocation DatesLocation

const webTitle string = "Groupie-Tracker"

func RenderTemplate(w http.ResponseWriter, tmpl string, page Page) {

	////////
	funcMap := template.FuncMap{
		"json": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
	}
	// include in map js func
	t, err := template.New(tmpl + ".html").Funcs(funcMap).ParseFiles("./web/templates/" + tmpl + ".html")

	if err != nil {
		ErrorPage(w, http.StatusBadRequest, Page{Title: webTitle, Error: "Oops, are you looking for a ghost..?"})
		return
	}

	if err := t.Execute(w, page); err != nil {
		ErrorPage(w, http.StatusBadRequest, Page{Title: webTitle, Error: "Oops, are you looking for a ghost..?"})
		return
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	//VarArtists = VarArtists[0:17]
	if VarArtists == nil {
		ErrorPage(w, http.StatusInternalServerError, Page{Title: webTitle, Error: "Oops, looks like the groupie API is empty..."})
		return
	}
	if r.URL.Path != "/" {
		idStr := string(r.URL.Path[1:])
		id, err := strconv.Atoi(idStr)
		id -= 1
		if err != nil || id < 0 || id > len(VarArtists)-1 {
			ErrorPage(w, http.StatusNotFound, Page{Title: webTitle, Error: "Oops, the page you were looking for could not be found... :)"})
			return
		} else {

			//////
			locations := make([]string, 0)
			for location := range VarDateLocation.Index[id].DatesLocations {
				locations = append(locations, location)
			}
			RenderTemplate(w, "Artist", Page{Title: webTitle, Data5: VarArtists[id], Data6: VarDateLocation.Index[id].DatesLocations, ArtistLocations: locations})
			return
		}
	}

	switch r.Method {
	case "GET":
		filters := getFiltersFromRequest(r)
		filteredArtists := applyFilters(VarArtists, VarDateLocation, filters)
		RenderTemplate(w, "Home", Page{Title: webTitle, Data1: filteredArtists, Filters: filters})
	default:
		ErrorPage(w, http.StatusMethodNotAllowed, Page{Title: webTitle, Error: "Oops, looks like you're not allowed to do this.."})
	}
}

func About(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/about" {
		ErrorPage(w, http.StatusNotFound, Page{Title: webTitle, Error: "Oops, the page you were looking for could not be found... :)"})
		return
	}

	switch r.Method {
	case "GET":
		RenderTemplate(w, "About", Page{Title: webTitle})
		log.Printf("HTTP Response Code : %v", (http.StatusOK))
	default:
		ErrorPage(w, http.StatusMethodNotAllowed, Page{Title: webTitle, Error: "Oops, looks like you're not allowed to do this.."})
	}
}

func ErrorPage(w http.ResponseWriter, errorCode int, page Page) {
	RenderTemplate(w, "Error", page)
	log.Printf("HTTP Response Code : %v", errorCode)
}

func getFiltersFromRequest(r *http.Request) Filters {
	return Filters{
		CreationDateFrom: getIntParam(r, "creationDate", 1950),
		CreationDateTo:   getIntParam(r, "creationDate", 2024),
		FirstAlbumFrom:   r.FormValue("first_album_from"),
		FirstAlbumTo:     r.FormValue("first_album_to"),
		MembersFrom:      getIntParam(r, "members_from", 1),
		MembersTo:        getIntParam(r, "members_to", 10),
		Locations:        getLocations(r.FormValue("locations")),
	}
}

func getIntParam(r *http.Request, name string, defaultValue int) int {
	value, err := strconv.Atoi(r.FormValue(name))
	if err != nil {
		return defaultValue
	}
	return value
}

func getLocations(locationsStr string) []string {
	if locationsStr == "" {
		return nil
	}
	return strings.Split(locationsStr, ",")
}

func applyFilters(artists []Artist, datesLocation DatesLocation, filters Filters) []Artist {
	var filteredArtists []Artist
	for i, artist := range artists {
		if i < len(datesLocation.Index) && matchesFilters(artist, datesLocation.Index[i], filters) {
			filteredArtists = append(filteredArtists, artist)
		}
	}
	return filteredArtists
}

func matchesFilters(artist Artist, dateLocation DateLocation, filters Filters) bool {
	// Vérifier la date de création seulement si les filtres sont spécifiés
	if filters.CreationDateFrom != 1950 || filters.CreationDateTo != 2024 {
		if artist.CreationDate < filters.CreationDateFrom || artist.CreationDate > filters.CreationDateTo {
			return false
		}
	}

	// Vérifier la date du premier album seulement si les filtres sont spécifiés
	if filters.FirstAlbumFrom != "" || filters.FirstAlbumTo != "" {
		firstAlbumDate, err := time.Parse("2006-01-02", artist.FirstAlbum)
		if err == nil {
			if filters.FirstAlbumFrom != "" {
				filterFromDate, err := time.Parse("2006-01-02", filters.FirstAlbumFrom)
				if err == nil && firstAlbumDate.Before(filterFromDate) {
					return false
				}
			}
			if filters.FirstAlbumTo != "" {
				filterToDate, err := time.Parse("2006-01-02", filters.FirstAlbumTo)
				if err == nil && firstAlbumDate.After(filterToDate) {
					return false
				}
			}
		}
	}

	// Vérifier le nombre de membres seulement si les filtres sont spécifiés
	if filters.MembersFrom != 1 || filters.MembersTo != 10 {
		if len(artist.Members) < filters.MembersFrom || len(artist.Members) > filters.MembersTo {
			return false
		}
	}

	// Vérifier les locations seulement si le filtre est spécifié
	if len(filters.Locations) > 0 {
		locationMatch := false
		for _, filterLocation := range filters.Locations {
			for location := range dateLocation.DatesLocations {
				if strings.Contains(strings.ToLower(location), strings.ToLower(filterLocation)) {
					locationMatch = true
					break
				}
			}
			if locationMatch {
				break
			}
		}
		if !locationMatch {
			return false
		}
	}

	return true
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query missing", http.StatusBadRequest)
		return
	}

	results := searchArtists(query, VarArtists, VarDateLocation)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func searchArtists(query string, artists []Artist, datesLocation DatesLocation) []map[string]interface{} {
	query = strings.ToLower(query)
	var results []map[string]interface{}

	for _, artist := range artists {
		// Recherche par nom d'artiste
		if strings.Contains(strings.ToLower(artist.Name), query) {
			results = append(results, createArtistResult(artist))
		}

		// Recherche par membre
		for _, member := range artist.Members {
			if strings.Contains(strings.ToLower(member), query) {
				results = append(results, createArtistResult(artist))
				break
			}
		}

		// Recherche par location
		if artist.Id < len(datesLocation.Index) {
			for location := range datesLocation.Index[artist.Id].DatesLocations {
				if strings.Contains(strings.ToLower(location), query) {
					results = append(results, createArtistResult(artist))
					break
				}
			}
		}

		// Recherche par date de premier album
		if strings.Contains(strings.ToLower(artist.FirstAlbum), query) {
			results = append(results, createArtistResult(artist))
		}

		// Recherche par date de création
		if strings.Contains(strings.ToLower(strconv.Itoa(artist.CreationDate)), query) {
			results = append(results, createArtistResult(artist))
		}
	}

	return results
}

func createArtistResult(artist Artist) map[string]interface{} {
	return map[string]interface{}{
		"type":  "artist/band",
		"id":    artist.Id,
		"name":  artist.Name,
		"image": artist.Image,
	}
}
