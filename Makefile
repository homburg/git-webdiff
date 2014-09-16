.PHONY: types

.SUFFIXES:

all: node_modules bower_components public/hello.js public/base.html public/vendor.js public/hello.css

run: git-webdiff all
	./git-webdiff develop

git-webdiff: $(shell find . -name "*.go")
	go build -o git-webdiff

node_modules:
	npm -q update

bower_components:
	bower -q update

public/hello.js: app/hello.ls
	gulp

public/%.html: app/%.jade
	gulp

public/vendor.js: bower.json
	gulp

public/hello.css: app/hello.styl
	gulp

setup:
	npm install -q -g bower gulp
