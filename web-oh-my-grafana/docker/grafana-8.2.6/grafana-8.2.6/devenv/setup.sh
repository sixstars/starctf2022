#!/bin/bash

bulkDashboard() {

		requiresJsonnet

		COUNTER=0
		MAX=400
		while [  $COUNTER -lt $MAX ]; do
				jsonnet -o "bulk-dashboards/dashboard${COUNTER}.json" -e "local bulkDash = import 'bulk-dashboards/bulkdash.jsonnet'; bulkDash + {  uid: 'uid-${COUNTER}',  title: 'title-${COUNTER}' }"
				let COUNTER=COUNTER+1
		done

		ln -s -f ../../../devenv/bulk-dashboards/bulk-dashboards.yaml ../conf/provisioning/dashboards/custom.yaml
}

bulkAlertingDashboard() {

		requiresJsonnet

		jsonnet -o "bulk_alerting_dashboards/bulk_alerting_datasources.yaml" ./bulk_alerting_dashboards/datasources.jsonnet

		COUNTER=1
		DS=1
		MAX=1000
		while [  $COUNTER -lt $MAX ]; do
				jsonnet -o "bulk_alerting_dashboards/alerting_dashboard${COUNTER}.json" \
					-e "local bulkDash = import 'bulk_alerting_dashboards/dashboard.libsonnet'; bulkDash.alertingDashboard(${COUNTER}, ${DS})"
				let COUNTER=COUNTER+1
				let DS=COUNTER/10
				let DS=DS+1
		done

		ln -s -f ../../../devenv/bulk_alerting_dashboards/bulk_alerting_dashboards.yaml ../conf/provisioning/dashboards/custom.yaml
		ln -s -f ../../../devenv/bulk_alerting_dashboards/bulk_alerting_datasources.yaml ../conf/provisioning/datasources/custom.yaml
}

requiresJsonnet() {
		if ! type "jsonnet" > /dev/null; then
				echo "you need you install jsonnet to run this script"
				echo "follow the instructions on https://github.com/google/jsonnet"
				exit 1
		fi
}

devDashboards() {
		echo -e "\xE2\x9C\x94 Setting up all dev dashboards using provisioning"
		ln -s -f ../../../devenv/dashboards.yaml ../conf/provisioning/dashboards/dev.yaml
}

devDatasources() {
		echo -e "\xE2\x9C\x94 Setting up all dev datasources using provisioning"

		ln -s -f ../../../devenv/datasources.yaml ../conf/provisioning/datasources/dev.yaml
}

usage() {
	echo -e "\n"
	echo "Usage:"
	echo "  bulk-dashboards                              - create and provisioning 400 dashboards"
	echo "  bulk-alerting-dashboards                     - create and provisioning 400 dashboards with alerts"
	echo "  no args                                      - provisioning core datasources and dev dashboards"
}

main() {
	echo -e "------------------------------------------------------------------"
	echo -e "This script sets up provisioning for dev datasources and dashboards"
	echo -e "------------------------------------------------------------------"
	echo -e "\n"

	local cmd=$1

	if [[ $cmd == "bulk-alerting-dashboards" ]]; then
		bulkAlertingDashboard
	elif [[ $cmd == "bulk-dashboards" ]]; then
		bulkDashboard
	else
		devDashboards
		devDatasources
	fi

  if [[ -z "$cmd" ]]; then
		usage
	fi

}

main "$@"
