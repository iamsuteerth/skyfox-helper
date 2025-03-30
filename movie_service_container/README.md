# movie-service

movie api to get movie details

**`get /movies` - returns list of all movies**

Example response:

```
[
    {
    "@type": "imdb.api.title.title",
    "id": "tt6137778",
    "image": {
      "height": 8748,
      "id": "tt6137778/images/rm4025404416",
      "url": "https://m.media-amazon.com/images/M/MV5BMTYyMjIyMjY4N15BMl5BanBnXkFtZTgwMzAxNTQ5MTE@._V1_.jpg",
      "width": 5906
    },
    "title": "The Well",
    "titleType": "movie",
    "year": 2020
  },
  {
    "@type": "imdb.api.title.title",
    "id": "tt6702810",
    "image": {
      "height": 2048,
      "id": "tt6702810/images/rm2582992384",
      "url": "https://m.media-amazon.com/images/M/MV5BMjI0MDMzNTQ0M15BMl5BanBnXkFtZTgwMTM5NzM3NDM@._V1_.jpg",
      "width": 1382
    },
    "runningTimeInMinutes": 90,
    "title": "A Quiet Place",
    "titleType": "movie",
    "year": 2018
  }

]
```

------------

**`get /movie/:id` - returns single movie details matching movie id**

```
{
    "@type": "imdb.api.title.title",
    "id": "tt6137778",
    "image": {
      "height": 8748,
      "id": "tt6137778/images/rm4025404416",
      "url": "https://m.media-amazon.com/images/M/MV5BMTYyMjIyMjY4N15BMl5BanBnXkFtZTgwMzAxNTQ5MTE@._V1_.jpg",
      "width": 5906
    },
    "title": "The Well",
    "titleType": "movie",
    "year": 2020
  }
```
