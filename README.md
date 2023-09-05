# Votes


this go project is designed to:
1) take messages that an MP might write
2) classify those messages into the different policy groups (defined by www.publicwhip.org.uk) usoing chatGPT
3) find the MP's voting record from www.publicwhip.org.uk 
4) build some sort of response highlighing their voting position on the related policy to post as a reply.

# running the project

`go run cmd/withgpt/main.go` will give you an idea of how it's supposed to work.
* the real version will need to run a microservice with an API that takes a tweet / post and returns the appropriate voting history message or an error

# different parts of the project
## data from PublicWhip
this unfortunately has to be scraped from their website
* that is largely what the `publicwhip` package does
however - getting a list of MPs along with their internal ID number takes time (and wont change often), so is currently cached in the file `mpData` - if you delete this file, next time it runs it will be regenerated. 


## members-api.parliament.uk
there is code that can talk to the official parliament api, but so far we have not needed this

## gpt
this is the package that talks to chatGPT and uses it to classify mesages using the policies from publicwhip - it was built quickly and could certainly do with some love
- also the prompt is probably suboptimal, but seems to work well enough for the moment.

