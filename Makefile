IMAGE   := ai-upskill
CONTAINER := ai-upskill

DATE ?= $(shell date +%Y-%m-%d)

.PHONY: build run stop logs shell clean local report podcast

build:
	docker build -t $(IMAGE) .

run:
	python3 -c "\
	import re;\
	out=open('.env.docker','w');\
	[out.write(re.sub(r\"^(\\w+=)(?:'(.+)'|\\\"(.+)\\\")\",r'\\1\\2\\3',re.sub(r'^export ','',l.strip()))+'\n') for l in open('.env') if l.strip() and not l.startswith('#')];\
	out.close()"
	docker run -d --name $(CONTAINER) \
		--env-file .env.docker \
		-v $(PWD)/data:/app/data \
		-v $(PWD)/reports:/app/reports \
		-v $(PWD)/podcasts:/app/podcasts \
		$(IMAGE) tail -f /dev/null
	rm -f .env.docker

stop:
	docker rm -f $(CONTAINER) 2>/dev/null || true

logs:
	docker logs -f $(CONTAINER)

shell:
	docker exec -it $(CONTAINER) bash

local: stop build run

report:
	docker exec $(CONTAINER) ./ai-report generate --date $(DATE)

podcast:
	docker exec $(CONTAINER) python3 scripts/generate-podcast.py start --date $(DATE) --media-type audio
	@echo "Polling for completion (typically 5-10 min)..."
	@attempt=0; while [ $$attempt -lt 40 ]; do \
		attempt=$$((attempt + 1)); \
		echo "Poll $$attempt/40..."; \
		if docker exec $(CONTAINER) python3 scripts/generate-podcast.py poll 2>&1; then \
			echo "Downloading..."; \
			docker exec $(CONTAINER) python3 scripts/generate-podcast.py download; \
			exit 0; \
		fi; \
		sleep 30; \
	done; \
	echo "Timed out after 20 minutes"; exit 1

clean:
	docker rm -f $(CONTAINER) 2>/dev/null || true
	docker rmi -f $(IMAGE) 2>/dev/null || true
