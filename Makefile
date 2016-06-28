.PHONY: default dev dist packaging packaging_destroy fmt test testv deps updatedeps

default: dev

dev:
	@bash -c $(CURDIR)/_build/dev.sh

dist:
	@bash -c $(CURDIR)/_build/dist.sh

packaging:
	@bash -c $(CURDIR)/_build/packaging.sh

packaging_destroy:
	@sh -c "cd $(CURDIR)/_build/packaging/rpm && vagrant destroy -f"

fmt:
	go fmt ./...

test:
	go test ./... -cover

testv:
	go test ./... -v

test_integration:
	cd _tests && \
    vagrant destroy -f && \
    vagrant up && \
	vagrant ssh centos-6 -c "cd /home/vagrant/src/github.com/kohkimakimoto/cofu/_tests && sudo bash run.sh" && \
    vagrant ssh centos-7 -c "cd /home/vagrant/src/github.com/kohkimakimoto/cofu/_tests && sudo bash run.sh"

deps:
	gom install

deps_update:
	rm Gomfile.lock; rm -rf _vendor; gom install && gom lock
