import React, { useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { notification } from 'antd';
import { magicLogin } from '../services/userService';

const MagicLogin: React.FC = () => {
    const [searchParams] = useSearchParams();

    useEffect(() => {
        const token = searchParams.get('token');
        if (token) {
            (async () => {
                try {
                    const data = await magicLogin(token);
                    localStorage.setItem('token', data.authToken); 
                    notification.success({
                        message: 'Login Successful',
                        description: 'You are now logged in.',
                    });
                    window.location.href = '/'; 
                } catch (error) {
                    notification.error({
                        message: 'Login Failed',
                        description: error instanceof Error ? error.message : 'Invalid or expired magic link.',
                    });
                }
            })();
        }
    }, [searchParams]);

    return <div>Logging you in...</div>;
};

export default MagicLogin;
