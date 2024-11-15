import React, { useState } from 'react';
import { Form, Input, Button, notification } from 'antd';
import { verifyCode } from '../services/userService';

const VerificationPage: React.FC = () => {
  const [code, setCode] = useState<string>('');
  const [loading, setLoading] = useState(false);

  const onFinish = async () => {
    setLoading(true);
    try {
      const response = await verifyCode(code);
      if (response.status === 200) {
        notification.success({
          message: 'Verification Successful',
          description: 'Your account has been successfully activated.',
        });
        window.location.href = '/login';
      }
    } catch (error) {
      notification.error({
        message: 'Error',
        description: 'The verification code is incorrect or expired.',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: 400, margin: '0 auto', padding: '2rem' }}>
      <h2>Enter Verification Code</h2>
      <Form onFinish={onFinish} layout="vertical">
        <Form.Item
          label="Verification Code"
          name="code"
          rules={[{ required: true, message: 'Please input your verification code!' }]}
        >
          <Input
            value={code}
            onChange={(e) => setCode(e.target.value)}
            maxLength={6}
            placeholder="Enter the 6-digit code"
          />
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading}>
            Verify
          </Button>
        </Form.Item>
      </Form>
    </div>
  );
};

export default VerificationPage;
