import React, { useState } from 'react';
import { Form, Input, Button, notification } from 'antd';
import { requestMagicLink } from '../services/userService';

const MagicLinkPage: React.FC = () => {
	const [email, setEmail] = useState<string>('');
	const [sendingMagicLink, setSendingMagicLink] = useState(false);

	const onRequestMagicLink = async () => {
		setSendingMagicLink(true);
		try {
			await requestMagicLink(email);
			notification.success({
				message: 'Magic Link Sent',
				description: 'A magic login link has been sent to your email.',
			});
		} catch (error) {
			notification.error({
				message: 'Error',
				description: 'Failed to send magic link. Please try again.',
			});
		} finally {
			setSendingMagicLink(false);
		}
	};

	return (
		<div style={{ maxWidth: 400, margin: '0 auto', padding: '2rem' }}>
			<h2>Request Magic Link</h2>
			<Form layout="vertical">
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
					<Button
						type="default"
						onClick={onRequestMagicLink}
						loading={sendingMagicLink}
						block
					>
						Send Magic Link
					</Button>
				</Form.Item>
			</Form>
		</div>
	);
};

export default MagicLinkPage;
