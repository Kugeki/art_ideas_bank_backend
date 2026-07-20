#!/bin/sh

env_file=".env"

if [ -f "$env_file" ]; then
  echo "Environment file \"$env_file\" already exists."
  read -r -p "Do you want to overwrite it? (y/N): " confirmation
  if [[ ! $confirmation =~ ^[Yy]$ ]]; then
    exit 1
  fi
  mv "$env_file" "${env_file}_old_$(date +%Y%m%d_%H%M%S)" # if updating name, also update .gitignore
fi

echo "Generating environment file \"$env_file\"..."

cat <<EOF > "$env_file"
# Generated on $(date)
MAIN_POSTGRES_USER=art_ideas_bank_user
MAIN_POSTGRES_PASSWORD=$(openssl rand -hex 32)
MAIN_POSTGRES_DB=art_ideas_bank

GARAGE_RPC_SECRET=$(openssl rand -hex 32)
GARAGE_ADMIN_TOKEN=$(openssl rand -base64 32)
GARAGE_METRICS_TOKEN=$(openssl rand -base64 32)

S3_ACCESS_KEY=GK$(openssl rand -hex 32)
S3_SECRET_KEY=$(openssl rand -hex 32)
S3_BUCKET=art-ideas-bank-bucket

JWT_SECRET_KEY=$(openssl rand -hex 32)
EOF

chmod 640 "$env_file"

echo "Environment file \"$env_file\" successfully generated."
