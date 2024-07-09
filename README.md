
# 📜 Groupie-tracker

web server handled in Golang 1.22.4.

This project contain all features of groupie tracker. 

- Filters
- Searchbar
- Vizualisation
- Geolocalization

## ✍🏼 Authors

- [@pcaboor](https://zone01normandie.org/git/pcaboor/)


## 📦 How it's work ?

Firstly you need to extract cites coordinations in a `coordinate.json`

Example : 

```json
{
    "north_carolina-usa": [-79.0193, 35.7596],
    "georgia-usa": [-82.9001, 32.1656],
    "los_angeles-usa": [-118.2437, 34.0522],
    "saitama-japan": [139.6489, 35.8617],
    "osaka-japan": [135.5022, 34.6937],
    "nagoya-japan": [136.9066, 35.1815],
    "penrose-new_zealand": [174.7691, -36.8667],
    "dunedin-new_zealand": [170.5036, -45.8788],
    "playa_del_carmen-mexico": [-87.0739, 20.6296],
    "papeete-french_polynesia": [-149.5665, -17.5516],
    "noumea-new_caledonia": [166.4505, -22.2711],
    "london-uk": [-0.1276, 51.5074],
    "lausanne-switzerland": [6.6323, 46.5197],
    "lyon-france": [4.8357, 45.7640],
    "victoria-australia": [144.9631, -37.8136],
    // other cites....
}
```

Please note that cites correspond to this API : https://groupietrackers.herokuapp.com/api/locations

Secondly I create a python script to convert this coordinations to `.geojson`

`geoson` is a `json` file that contains more informations about a locations, this is the struct :

```json
 {
      "type": "Feature",
      "geometry": {
        "type": "Point",
        "coordinates": [
          -123.1216,
          49.2827
        ]
      },
      "properties": {
        "name": "vancouver-canada"
      }
    },
```   

Now I need to use a service call [MapTiler](https://cloud.maptiler.com/) to display the map with pointers.

I import my `data.geojson` to create a datasets ( place pointer on the map wiht coordinates ).

After that I have a API key for my map and pointers.

In my code i call my API in my HTML page in script :

```javascript
 <script>
        var artistLocationsData = '{{.ArtistLocations | json}}';
        var artistLocations = JSON.parse(artistLocationsData);

        const apiKey = '3vpA40hTYfVxxZHYLSk9';
        
    //    const dataUrl = 'https://api.maptiler.com/data/67d61f8f-dff3-4126-9c43-c5e011c4ce89/features.json?key=3vpA40hTYfVxxZHYLSk9';

        const map = new maplibregl.Map({
            container: 'map',
            style: 'https://api.maptiler.com/maps/ch-swisstopo-lbm-dark/style.json?key=3vpA40hTYfVxxZHYLSk9',
            center: [-0.07, 51.5],
            zoom: 4,
            hash: true
        });

        map.on('load', function () {
            fetch(dataUrl)
                .then(response => response.json())
                .then(data => {
                    const filteredFeatures = data.features.filter(feature =>
                        artistLocations.includes(feature.properties.name)
                    );

                    map.addSource('points', {
                        type: 'geojson',
                        data: {
                            type: 'FeatureCollection',
                            features: filteredFeatures
                        }
                    });
                  
                    map.addLayer({
                        id: 'points',
                        type: 'circle',
                        source: 'points',
                        paint: {
                            'circle-radius': 6,
                            'circle-color': '#00FF00'
                        }
                    });
                });
        });

        map.on('error', function (err) {
            console.error('An error occurred:', err);
        });
    </script>
    
 ```

Finally in the `pageHandler.go` :

```go
funcMap := template.FuncMap{
		"json": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
	}
	// include in map js func
	t, err := template.New(tmpl + ".html").Funcs(funcMap).ParseFiles("./web/templates/" + tmpl + ".html")
```

It's very important because without that, I can't link data from artist to location in MapTiler 

and :

```go
	locations := make([]string, 0)
			for location := range VarDateLocation.Index[id].DatesLocations {
				locations = append(locations, location)
			}
```
         