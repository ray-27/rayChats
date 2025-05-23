import json
import smtplib
import random
import string
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
import os
from dotenv import load_dotenv

load_dotenv()

def lambda_handler(event, context):
    # Gmail SMTP configuration
    gmail_user = os.environ.get('GMAIL_USER')  # Your Gmail address
    gmail_app_password = os.environ.get('GMAIL_APP_PASSWORD')  # Your app password
    
    print(f"Gmail User: {gmail_user}")
    print(f"Password exists: {bool(gmail_app_password)}")

    try:
        # Parse the request body
        if isinstance(event.get('body'), str):
            body = json.loads(event['body'])
        else:
            body = event.get('body', {})
        
        # Get email from request
        email = body.get('email')
        
        if not email:
            return {
                'statusCode': 400,
                'headers': {
                    'Access-Control-Allow-Origin': '*',
                    'Access-Control-Allow-Headers': 'Content-Type',
                    'Access-Control-Allow-Methods': 'POST, OPTIONS'
                },
                'body': json.dumps({
                    'success': False,
                    'error': 'Email is required'
                })
            }
        
        # Generate 6-digit OTP
        otp = ''.join(random.choices(string.digits, k=6))
        
        
        
        if not gmail_user or not gmail_app_password:
            return {
                'statusCode': 500,
                'headers': {
                    'Access-Control-Allow-Origin': '*',
                    'Access-Control-Allow-Headers': 'Content-Type',
                    'Access-Control-Allow-Methods': 'POST, OPTIONS'
                },
                'body': json.dumps({
                    'success': False,
                    'error': 'Gmail credentials not configured'
                })
            }
        
        # HTML email template
        html_body = f"""
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <title>Your OTP Code</title>
        </head>
        <body style="font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px;">
            <div style="max-width: 600px; margin: 0 auto; background-color: white; border-radius: 10px; padding: 40px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
                <div style="text-align: center;">
                    <h2 style="color: #333; margin-bottom: 30px;"> rayChatApp Email Verification</h2>
                    <p style="color: #666; font-size: 16px; margin-bottom: 30px;">
                        Please use the following OTP to verify your email address:
                    </p>
                    
                    <!-- OTP in large heading -->
                    <h1 style="color: #7D56F4; font-size: 48px; letter-spacing: 8px; margin: 30px 0; padding: 20px; background-color: #f8f9fa; border-radius: 8px; border: 2px dashed #7D56F4;">
                        {otp}
                    </h1>
                    
                    <p style="color: #666; font-size: 14px; margin-top: 30px;">
                        This OTP will expire in 10 minutes.
                    </p>
                    <p style="color: #999; font-size: 12px; margin-top: 20px;">
                        If you didn't request this verification, please ignore this email.
                    </p>
                    
                    <div style="margin-top: 40px; padding-top: 20px; border-top: 1px solid #eee;">
                        <p style="color: #999; font-size: 12px;">
                            © 2025 ChatApp. All rights reserved.
                        </p>
                    </div>
                </div>
            </div>
        </body>
        </html>
        """
        
        # Create message
        msg = MIMEMultipart('alternative')
        msg['Subject'] = 'Your rayChatApp Verification Code'
        msg['From'] = gmail_user
        msg['To'] = email
        
        # Create HTML part
        html_part = MIMEText(html_body, 'html')
        
        # Create plain text part
        text_body = f"""
        rayChatApp Email Verification
        
        Your verification code is: {otp}
        
        This code will expire in 10 minutes.
        
        If you didn't request this verification, please ignore this email.
        """
        text_part = MIMEText(text_body, 'plain')
        
        # Add parts to message
        msg.attach(text_part)
        msg.attach(html_part)
        
        # Send email using Gmail SMTP
        try:
            server = smtplib.SMTP('smtp.gmail.com', 587)
            server.starttls()  # Enable encryption
            server.login(gmail_user, gmail_app_password)
            
            text = msg.as_string()
            server.sendmail(gmail_user, email, text)
            server.quit()
            
            # Return success response
            return {
                'statusCode': 200,
                'headers': {
                    'Access-Control-Allow-Origin': '*',
                    'Access-Control-Allow-Headers': 'Content-Type',
                    'Access-Control-Allow-Methods': 'POST, OPTIONS'
                },
                'body': json.dumps({
                    'success': True,
                    'message': 'OTP sent successfully',
                    'otp': otp  
                })
            }
            
        except Exception as smtp_error:
            return {
                'statusCode': 500,
                'headers': {
                    'Access-Control-Allow-Origin': '*',
                    'Access-Control-Allow-Headers': 'Content-Type',
                    'Access-Control-Allow-Methods': 'POST, OPTIONS'
                },
                'body': json.dumps({
                    'success': False,
                    'error': f'Failed to send email: {str(smtp_error)}'
                })
            }
        
    except Exception as e:
        return {
            'statusCode': 500,
            'headers': {
                'Access-Control-Allow-Origin': '*',
                'Access-Control-Allow-Headers': 'Content-Type',
                'Access-Control-Allow-Methods': 'POST, OPTIONS'
            },
            'body': json.dumps({
                'success': False,
                'error': str(e)
            })
        }

