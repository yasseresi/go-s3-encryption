# 🔐 Cloud Cryptography Lab

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/)
[![AWS SDK v2](https://img.shields.io/badge/AWS%20SDK-v2-FF9900?style=for-the-badge&logo=amazonaws&logoColor=white)](https://aws.amazon.com/sdk-for-go/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](LICENSE)

**An educational project demonstrating cloud storage encryption patterns**

[🚀 Quick Start](#-quick-start) • [📖 Documentation](#-how-it-works) • [🔧 Configuration](#-configuration) • [🛡️ Security](#-security-notes)

</div>

---

## 🎯 Overview

**Cloud Cryptography Lab** is an educational web application designed for Computer Science students and security enthusiasts. It provides hands-on experience with two fundamental patterns of cloud storage security using AWS S3 and Go:

- **🔒 Server-Side Encryption (SSE-KMS)** - AWS-managed encryption
- **🛡️ Client-Side Encryption (CSE-AES-GCM)** - Application-controlled encryption

The application features a modern, dark-themed web interface that visualizes cryptographic concepts and allows you to experiment with different encryption strategies in a real cloud environment.

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🔒 Server-Side Encryption
- **Transparent encryption** using AWS KMS
- **Automatic key management** by AWS
- **Zero-knowledge uploads** - plaintext over TLS
- **Seamless decryption** on download

</td>
<td width="50%">

### 🛡️ Client-Side Encryption  
- **AES-256-GCM** encryption before upload
- **Customer-managed keys** (base64 encoded)
- **Zero-trust architecture** - AWS never sees plaintext
- **Local decryption** with ciphertext preview

</td>
</tr>
</table>

### 🌐 Educational Interface
- **Visual feedback** showing plaintext vs ciphertext
- **Object browser** with `sse/` and `cse/` prefixes
- **Real-time encryption status** indicators
- **Modern dark theme** optimized for developers

---

## 🚀 Quick Start

### Prerequisites
- **Go 1.23+** installed ([Download here](https://golang.org/dl/))
- **AWS Account** with S3 and KMS access
- **AWS Credentials** configured locally

### Installation

\`\`\`bash
# 1. Clone the repository
git clone https://github.com/yourusername/cloud-cryptography-lab.git
cd cloud-cryptography-lab

# 2. Install dependencies
go mod tidy

# 3. Configure environment (see Configuration section)
cp .env.example .env
# Edit .env with your AWS credentials

# 4. Run the application
go run .

# 5. Open your browser
open http://localhost:8080
\`\`\`

---

## 🔧 Configuration

Create a `.env` file in the root directory:

\`\`\`bash
# 🌍 AWS Configuration
AWS_REGION=us-east-1
S3_BUCKET=your-unique-bucket-name
AWS_ACCESS_KEY_ID=your_access_key_here
AWS_SECRET_ACCESS_KEY=your_secret_key_here

# 🔑 KMS Key for Server-Side Encryption
SSE_KMS_KEY_ID=arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012

# 🛡️ Customer Key for Client-Side Encryption (32 bytes, base64 encoded)
# Generate with: openssl rand -base64 32
CUSTOMER_KEY_BASE64=YourGeneratedBase64KeyHere==

# 🌐 Server Configuration
PORT=8080
\`\`\`

> **⚠️ Security Note:** Never commit your `.env` file to version control!

---

## 📖 How It Works

### 🔒 Server-Side Encryption (SSE-KMS) Flow

\`\`\`
📄 Plaintext File → 🚀 HTTPS Upload → ☁️ AWS S3 → 🔑 KMS Encryption → 💾 Encrypted Storage
\`\`\`

1. **Upload**: File sent as plaintext over HTTPS
2. **Storage**: S3 encrypts using your KMS key
3. **Download**: S3 auto-decrypts if you have permissions

### 🛡️ Client-Side Encryption (CSE-AES-GCM) Flow

\`\`\`
📄 Plaintext File → 🔐 AES-GCM Encrypt → 🚀 Ciphertext Upload → ☁️ AWS S3 → 💾 Ciphertext Storage
\`\`\`

1. **Encrypt**: AES-256-GCM encryption locally
2. **Upload**: Only ciphertext sent to S3
3. **Download**: Decrypt locally with customer key

---

## 🛡️ Security Notes

| ⚠️ **Important** | This is an educational project |
|------------------|-------------------------------|
| 🎓 **Purpose** | Demonstration and learning only |
| 🔑 **Key Management** | Use proper vaults in production (AWS Secrets Manager, HashiCorp Vault) |
| 🔄 **Key Rotation** | Implement rotation strategies for production |
| 👤 **IAM Permissions** | Use least-privilege access |

### Required IAM Permissions

\`\`\`json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject", 
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::your-bucket-name",
        "arn:aws:s3:::your-bucket-name/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt",
        "kms:GenerateDataKey"
      ],
      "Resource": "arn:aws:kms:region:account:key/key-id"
    }
  ]
}
\`\`\`

---

## 📁 Project Structure

\`\`\`
cloud-cryptography-lab/
├── 🚀 main.go              # Application entry point & config loading
├── 🌐 handlers.go          # HTTP handlers (Upload, List, Download)  
├── ☁️ s3.go               # AWS S3 SDK integration & utilities
├── 🔐 crypto.go           # AES-GCM encryption/decryption logic
├── 📁 templates/
│   └── 🎨 index.html      # Modern web interface (HTML/CSS)
├── 🔧 go.mod              # Go module dependencies
├── 🙈 .gitignore          # Git ignore rules
├── 📋 .env.example        # Environment template
└── 📖 README.md           # This documentation
\`\`\`

---

## 🎨 Tech Stack

<div align="center">

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Backend** | ![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go) | Core application logic |
| **Cloud** | ![AWS](https://img.shields.io/badge/AWS-S3%20%7C%20KMS-FF9900?style=flat&logo=amazonaws) | Storage & key management |
| **Crypto** | ![AES](https://img.shields.io/badge/AES--256--GCM-Standard%20Library-green?style=flat) | Client-side encryption |
| **Frontend** | ![HTML5](https://img.shields.io/badge/HTML5-CSS3-E34F26?style=flat&logo=html5) | Modern dark UI |
| **Config** | ![dotenv](https://img.shields.io/badge/.env-Configuration-yellow?style=flat) | Environment management |

</div>

---

## 📝 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ for Computer Science Education**

⭐ **Star this repo if it helped you learn about cloud cryptography!** ⭐

</div>