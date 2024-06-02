from openai import OpenAI
client = OpenAI()

response = client.images.generate(
  model="dall-e-3",
  prompt="the Palace of Westminster on a clear day from the bank across the river in a cartoony style",
  size="1792x1024",
  quality="standard",
  n=1,
)

image_url = response.data[0].url

print(image_url)