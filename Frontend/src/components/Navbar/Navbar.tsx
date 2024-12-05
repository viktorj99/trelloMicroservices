import React, { useState } from 'react';
import { Menu, Dropdown, Button, Drawer, Spin, List, Badge } from 'antd';
import {
	HomeOutlined,
	UserOutlined,
	SettingOutlined,
	LogoutOutlined,
	BellOutlined,
} from '@ant-design/icons';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { removeToken, isUserLoggedIn, isManager, getTokenData } from '../../utils/authHelpers';
import { fetchNotifications } from '../../services/notificationService';

const Navbar: React.FC = () => {
	const tokenData = getTokenData();
	const [drawerVisible, setDrawerVisible] = useState(false);
	const [notifications, setNotifications] = useState<any[]>([]);
	const [loading, setLoading] = useState(false);

	const loadNotifications = async () => {
		if (!tokenData || !tokenData.id) return;
		setLoading(true);
		try {
			const data = await fetchNotifications(tokenData.id);
			setNotifications(data);
		} catch (error) {
			console.error('Failed to fetch notifications:', error);
		} finally {
			setLoading(false);
		}
	};

	// Open the drawer and load notifications
	const openDrawer = () => {
		setDrawerVisible(true);
		loadNotifications();
		console.log('Token Data: ', tokenData);
	};

	// Close the drawer
	const closeDrawer = () => {
		setDrawerVisible(false);
	};
	const navigate = useNavigate();
	const location = useLocation();

	const selectedKey =
		location.pathname === '/'
			? 'home'
			: location.pathname === '/projects'
			? 'projects'
			: location.pathname === '/project/create'
			? 'projectCreate'
			: null;

	const handleLogout = () => {
		removeToken();
		navigate('/login');
	};

	const guestMenu = (
		<Menu>
			<Menu.Item key='login'>
				<Link to='/login'>Login</Link>
			</Menu.Item>
			<Menu.Item key='registration'>
				<Link to='/registration'>Registration</Link>
			</Menu.Item>
		</Menu>
	);

	return (
		<div
			style={{
				backgroundColor: '#001529',
				display: 'flex',
				alignItems: 'center',
				padding: '0',
			}}
		>
			<Menu
				mode='horizontal'
				theme='dark'
				selectedKeys={selectedKey ? [selectedKey] : []}
				style={{
					flex: 1,
					borderBottom: 'none',
					justifyContent: 'flex-start',
				}}
			>
				<Menu.Item key='home' icon={<HomeOutlined />}>
					<Link to='/'>Home</Link>
				</Menu.Item>
				{isUserLoggedIn() && (
					<>
						<Menu.Item key='projects' icon={<SettingOutlined />}>
							<Link to='/projects'>Projects</Link>
						</Menu.Item>
						<Menu.Item key='profile' icon={<SettingOutlined />}>
							<Link to='/profile'>Profile</Link>
						</Menu.Item>
						<Menu.Item  key='notifications' icon={<BellOutlined />} onClick={openDrawer}>
							<Badge
								count={
									notifications.filter((n: { is_read: any }) => !n.is_read).length
								}
							>
								<span style={{color: "rgba(255, 255, 255, 0.65)"}}>Notifications</span>
								
							</Badge>
						</Menu.Item>
						<Drawer
							title='Notifications'
							placement='right'
							onClose={closeDrawer}
							open={drawerVisible}
							width={350}
						>
							{loading ? (
								<Spin />
							) : (
								<List
									dataSource={notifications}
									renderItem={(item) => (
										<List.Item>
											<List.Item.Meta
												title={item.message}
												description={new Date(
													item.created_at
												).toLocaleString()}
											/>
										</List.Item>
									)}
								/>
							)}
						</Drawer>
					</>
				)}
				{isManager() && (
					<>
						<Menu.Item key='projectCreate' icon={<SettingOutlined />}>
							<Link to='/project/create'>Create Project</Link>
						</Menu.Item>
					</>
				)}
			</Menu>

			<div style={{ display: 'flex', alignItems: 'center' }}>
				{isUserLoggedIn() ? (
					<Button
						type='text'
						icon={<LogoutOutlined />}
						style={{ color: '#fff' }}
						onClick={handleLogout}
					>
						Logout
					</Button>
				) : (
					<Dropdown overlay={guestMenu} placement='bottomRight'>
						<Button
							icon={<UserOutlined />}
							style={{
								color: 'white',
								backgroundColor: 'transparent',
								border: 'none',
							}}
						>
							Options
						</Button>
					</Dropdown>
				)}
			</div>
		</div>
	);
};

export default Navbar;
