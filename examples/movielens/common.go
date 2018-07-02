package movielens

import (
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/lovoo/cofire"
)

// ReadRatings reads a Movie Lens file and returns a slice of ratings
func ReadRatings(fname string) []cofire.Rating {
	dat, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	var ratings []cofire.Rating
	for _, l := range strings.Split(string(dat), "\n") {
		if l == "" {
			continue
		}
		e := strings.Split(l, ",")
		if len(e) != 4 {
			log.Print(l)
			log.Fatal("!= 4")
		}

		s, _ := strconv.ParseFloat(e[2], 64)
		ratings = append(ratings, cofire.Rating{
			UserId:    e[0],
			ProductId: e[1],
			Score:     s,
		})
	}

	rand.Shuffle(len(ratings), func(i, j int) {
		ratings[i], ratings[j] = ratings[j], ratings[i]
	})
	return ratings
}
