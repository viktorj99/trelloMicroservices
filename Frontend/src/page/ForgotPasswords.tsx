import React, { useState } from 'react';
import { Form, Input, Button, notification } from 'antd';
import { sendResetPasswordEmail } from '../services/userService';

const ForgotPasswordPage: React.FC = () => {
	const [email, setEmail] = useState<string>('');
	const [loading, setLoading] = useState(false);

	const onFinish = async () => {
		setLoading(true);
		try {
			await sendResetPasswordEmail(email);
			notification.success({
				message: 'Email Sent',
				description: 'A password reset link has been sent to your email.',
			});
			setTimeout(() => {
				window.location.href = '/change-password'; 
			}, 1000);
		} catch (error) {
			notification.error({
				message: 'Error',
				description: 'User with this email does not exist.',
			});
		} finally {
			setLoading(false);
		}
	};

	return (
		<div style={{ maxWidth: 400, margin: '0 auto', padding: '2rem' }}>
			<h2>Forgot Password</h2>
			<Form onFinish={onFinish} layout="vertical">
				<Form.Item
					label="Email Address"
					name="email"
					rules={[
						{ required: true, message: 'Please input your email!' },
						{ type: 'email', message: 'Please enter a valid email address!' },
					]}
				>
					<Input
						value={email}
						onChange={(e) => setEmail(e.target.value)}
						placeholder="Enter your email"
					/>
				</Form.Item>

				<Form.Item>
					<Button type="primary" htmlType="submit" loading={loading} block>
						Submit
					</Button>
				</Form.Item>
			</Form>
		</div>
	);
};

export default ForgotPasswordPage;
