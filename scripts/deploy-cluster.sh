#!/bin/bash

# Параметры развертывания
CLUSTER_NAME="raft-cluster"
REGION="us-east-1"
INSTANCE_TYPE="t3.medium"
NODE_COUNT=3
KEY_NAME="raft-cluster-key"
SECURITY_GROUP="raft-cluster-sg"

# Создание ключевой пары
echo "Creating key pair..."
aws ec2 create-key-pair --key-name ${KEY_NAME} --query 'KeyMaterial' --output text > ${KEY_NAME}.pem
chmod 400 ${KEY_NAME}.pem

# Создание security group
echo "Creating security group..."
GROUP_ID=$(aws ec2 create-security-group \
  --group-name ${SECURITY_GROUP} \
  --description "Security group for Raft cluster" \
  --query 'GroupId' --output text)

aws ec2 authorize-security-group-ingress \
  --group-id ${GROUP_ID} \
  --protocol tcp \
  --port 22 \
  --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
  --group-id ${GROUP_ID} \
  --protocol tcp \
  --port 8080-8082 \
  --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
  --group-id ${GROUP_ID} \
  --protocol tcp \
  --port 9090-9092 \
  --cidr 0.0.0.0/0

# Запуск инстансов
echo "Launching ${NODE_COUNT} instances..."
INSTANCE_IDS=$(aws ec2 run-instances \
  --image-id ami-0c02fb55956c7d316 \
  --count ${NODE_COUNT} \
  --instance-type ${INSTANCE_TYPE} \
  --key-name ${KEY_NAME} \
  --security-group-ids ${GROUP_ID} \
  --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=${CLUSTER_NAME}-node}]" \
  --query 'Instances[*].InstanceId' \
  --output text | tr '\t' ' ')

# Ожидание запуска инстансов
echo "Waiting for instances to start..."
aws ec2 wait instance-running --instance-ids ${INSTANCE_IDS}

# Получение публичных IP
PUBLIC_IPS=$(aws ec2 describe-instances \
  --instance-ids ${INSTANCE_IDS} \
  --query 'Reservations[*].Instances[*].PublicIpAddress' \
  --output text | tr '\t' ' ')

# Настройка кластера
echo "Configuring cluster nodes..."
i=1
for IP in ${PUBLIC_IPS}; do
  echo "Configuring node ${i} at ${IP}..."
  
  scp -i ${KEY_NAME}.pem bin/raft-kv-store-linux-amd64 ubuntu@${IP}:/home/ubuntu/raft-kv-store
  scp -i ${KEY_NAME}.pem configs/cluster.yaml ubuntu@${IP}:/home/ubuntu/
  
  ssh -i ${KEY_NAME}.pem ubuntu@${IP} <<EOF
    chmod +x /home/ubuntu/raft-kv-store
    sudo mv /home/ubuntu/raft-kv-store /usr/local/bin/
    sudo mkdir -p /etc/raft-kv-store
    sudo mv /home/ubuntu/cluster.yaml /etc/raft-kv-store/
EOF

  i=$((i+1))
done

echo "Cluster deployment completed"
echo "Nodes public IPs: ${PUBLIC_IPS}"