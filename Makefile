.PHONY: ping
ping:
	cd ansible && ansible -i brokers all -m ping

.PHONY: run-kafka
run-kafka:
	cd ansible && vagrant up && ansible-playbook kafka-playbook.yaml
