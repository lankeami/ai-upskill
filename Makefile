IMAGE   := ai-upskill
CONTAINER := ai-upskill

.PHONY: build run stop logs shell clean

build:
	docker build -t $(IMAGE) .

run:
	docker run -d --name $(CONTAINER) \
		--env-file .env \
		-v $(PWD)/data:/app/data \
		-v $(PWD)/reports:/app/reports \
		-v $(PWD)/podcasts:/app/podcasts \
		$(IMAGE) tail -f /dev/null

stop:
	docker rm -f $(CONTAINER) 2>/dev/null || true

logs:
	docker logs -f $(CONTAINER)

shell:
	docker exec -it $(CONTAINER) bash

clean:
	docker rm -f $(CONTAINER) 2>/dev/null || true
	docker rmi -f $(IMAGE) 2>/dev/null || true
