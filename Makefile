out/chatbot:
	go build -o $@ chatbot.go 

.PHONY: chatbot-image
chatbot-image: out/chatbot
	docker build -t gcr.io/priya-wadhwa/chatbot:latest -f Dockerfile .

.PHONY: push-chatbot-image
push-chatbot-image: chatbot-image
	docker push gcr.io/priya-wadhwa/chatbot:latest