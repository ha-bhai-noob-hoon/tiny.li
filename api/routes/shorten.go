package routes

import (
	"os"
	"strconv"
	"time"

	"example.com/database"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

type request struct {
	URL				string				`json:"url"`
	CustomShort		string				`json:"short"`
	Expiry			time.Duration		`json:"expiry"`
}

type response struct{
	URL				string				`json:"url"`
	CustomShort		string				`json:"short"`
	Expiry			time.Duration		`json:"expiry"`
	XRateRemaining	int					`json:"rate_limit"`
	XRateLimitRest	time.Duration		`json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {

	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"cannot parse JSON"})
	}

	//implement rate limiting 
	
	r2 := database.CreateClient(1)
	defer r2.Close()
	val, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		_ = r2.Set(database.Ctx , c.IP() , os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else{
		val, _ := r2.Get(database.Ctx , c.IP()).Result()
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx , c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "rate limit exccedded",
				"rate_limit_rest": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	//check of the input is an actual url

	if !govalidator.IsURL(body.URL){
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Invalid URL"})
	}

	// check for domain error
	if !helper.RemoveDomainError(body.URL){
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error":"you cant access this :)"})
	} 

	//enforce https, SSL
	body.URL = helper.EnforceHTTP(body.URL)

	r2.Decr(database.Ctx , c.IP())



}