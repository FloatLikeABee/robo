User Authorization and Authentication System Design

Overview

This document outlines the design for a user authorization and authentication system that will integrate with the existing micro-service system. The system will handle user registration, login, and authorization controls, including role-based access to different services and default channel IDs for in-system broadcasting.

System Components

1. User Authentication Service

Features:

User registration with email, username, and password
Email verification
Google OAuth integration
Password reset functionality
Session management
Technical Specifications:

Framework: Node.js with Express.js
Database: MongoDB for user data storage
Authentication: JWT (JSON Web Tokens)
Password Hashing: bcrypt
Email Service: Nodemailer for sending verification and reset emails
2. User Authorization Service

Features:

Role-based access control (RBAC)
Permission configurations for different services
Default channel ID assignment for in-system broadcasting
Technical Specifications:

Framework: Node.js with Express.js
Database: MongoDB for role and permission data storage
Authorization: JWT validation and role checking
3. Integration with Existing Services

Main Panel App

Access Control: Users with the "Main Panel" role can access the AI chat app and basic data viewing ebook platform.
Forms App

Access Control: Users with the "Forms" role can create and broadcast forms via email or links.
Email Composer

Access Control: Users with the "Email Composer" role can compose emails with customized data and templates.
Data Reports Display

Access Control: Users with the "Sharp Reports" role can generate and display data reports in PDF or Excel formats.
Message Broadcasting

Default Channel ID: Each user will have a default channel ID for receiving real-time messages.
4. Network Gateway Integration

Features:

API Gateway for routing requests to the appropriate services
Message broker integration for in-system broadcasting
Technical Specifications:

Framework: Node.js with Express.js
API Gateway: Express.js middleware for routing
Message Broker: RabbitMQ for in-system broadcasting
Database Schema

User Collection

json


{
  "_id": ObjectId,
  "email": String,
  "username": String,
  "password": String,
  "googleId": String,
  "isVerified": Boolean,
  "roles": [String],
  "defaultChannelId": String,
  "createdAt": Date,
  "updatedAt": Date
}
Role Collection

json


{
  "_id": ObjectId,
  "name": String,
  "permissions": [String],
  "createdAt": Date,
  "updatedAt": Date
}
API Endpoints

User Authentication Service

Register User

Endpoint: POST /api/auth/register
Request Body:
json


{
  "email": "user@example.com",
  "username": "username",
  "password": "password"
}
Response:
json


{
  "message": "User registered successfully. Please check your email for verification."
}
Verify Email

Endpoint: GET /api/auth/verify-email?token=<verification_token>
Response:
json


{
  "message": "Email verified successfully."
}
Login User

Endpoint: POST /api/auth/login
Request Body:
json


{
  "email": "user@example.com",
  "password": "password"
}
Response:
json


{
  "token": "jwt_token",
  "user": {
    "email": "user@example.com",
    "username": "username",
    "roles": ["Main Panel", "Forms"],
    "defaultChannelId": "channel_id"
  }
}
Google OAuth

Endpoint: GET /api/auth/google
Response: Redirects to Google OAuth consent screen
Google OAuth Callback

Endpoint: GET /api/auth/google/callback
Response: Redirects to the main panel app with JWT token
Forgot Password

Endpoint: POST /api/auth/forgot-password
Request Body:
json


{
  "email": "user@example.com"
}
Response:
json


{
  "message": "Password reset email sent."
}
Reset Password

Endpoint: POST /api/auth/reset-password
Request Body:
json


{
  "token": "reset_token",
  "password": "new_password"
}
Response:
json


{
  "message": "Password reset successfully."
}
User Authorization Service

Get User Roles

Endpoint: GET /api/auth/roles
Headers: Authorization: Bearer <jwt_token>
Response:
json


{
  "roles": ["Main Panel", "Forms"]
}
Get User Permissions

Endpoint: GET /api/auth/permissions
Headers: Authorization: Bearer <jwt_token>
Response:
json


{
  "permissions": ["create_form", "broadcast_form"]
}
Integration with Existing Services

Main Panel App

Middleware: Check for "Main Panel" role in JWT token
Endpoint: GET /api/main-panel
Headers: Authorization: Bearer <jwt_token>
Forms App

Middleware: Check for "Forms" role in JWT token
Endpoint: POST /api/forms
Headers: Authorization: Bearer <jwt_token>
Email Composer

Middleware: Check for "Email Composer" role in JWT token
Endpoint: POST /api/email/compose
Headers: Authorization: Bearer <jwt_token>
Data Reports Display

Middleware: Check for "Sharp Reports" role in JWT token
Endpoint: GET /api/reports
Headers: Authorization: Bearer <jwt_token>
Message Broadcasting

Default Channel ID

Assignment: When a user registers, a default channel ID is generated and assigned.
Endpoint: GET /api/auth/user
Headers: Authorization: Bearer <jwt_token>
Response:
json


{
  "user": {
    "email": "user@example.com",
    "username": "username",
    "roles": ["Main Panel", "Forms"],
    "defaultChannelId": "channel_id"
  }
}
Message Broker Integration

Publisher: Services publish messages to the default channel ID.
Subscriber: Users subscribe to their default channel ID to receive real-time messages.
Security Considerations

Password Hashing

Algorithm: bcrypt
Salt Rounds: 10
JWT Token

Secret Key: Environment variable
Expiration Time: 1 hour
Email Verification

Token Expiration: 24 hours
Password Reset

Token Expiration: 1 hour