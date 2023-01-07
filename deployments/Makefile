CP = cp

RELEASE_VERSION ?= latest
CHART_SEMANTIC_RELEASE_VERSION ?= $(shell git tag | sort -V | tail -1)
CHART_NAME=github-releases-notifier
CHART_PATH ?= helm/
CHART_RENDERED_TEMPLATES_TMP=../target/charts
HELM_CLI=helm

helm_clean:
	rm -rf $(CHART_RENDERED_TEMPLATES_TMP)
	rm -f $(CHART_NAME)-*.tgz

helm_lint:
	# Linting Helm Chart
	$(HELM_CLI) lint --debug $(CHART_PATH)

helm_template:
	# Generating rendered template files to: $(CHART_RENDERED_TEMPLATES_TMP)
	mkdir -p $(CHART_RENDERED_TEMPLATES_TMP)
	$(HELM_CLI) template $(CHART_NAME) --output-dir $(CHART_RENDERED_TEMPLATES_TMP) $(CHART_PATH)

helm_pkg: helm_lint
	# Copying unarchived Helm Chart to release directory
	mkdir -p $(CHART_NAME)-$(RELEASE_VERSION)/charts/ 
	$(CP) -r $(CHART_PATH) $(CHART_NAME)-$(RELEASE_VERSION)/charts/$(CHART_NAME)
	# Packaging helm chart with semantic version: $(CHART_SEMANTIC_RELEASE_VERSION)
	$(HELM_CLI) package --version $(CHART_SEMANTIC_RELEASE_VERSION) --app-version $(CHART_SEMANTIC_RELEASE_VERSION) --destination ./ $(CHART_PATH)
	rm -rf $(CHART_NAME)-$(RELEASE_VERSION)

all: helm_clean helm_lint helm_template helm_pkg
template: helm_lint helm_template
clean: helm_clean

.PHONY: helm_clean helm_lint helm_template