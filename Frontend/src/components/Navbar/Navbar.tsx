import React from 'react';
import { Menu, Dropdown, Button } from 'antd';
import { HomeOutlined, UserOutlined, SettingOutlined, LogoutOutlined } from '@ant-design/icons';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { removeToken, isUserLoggedIn, isManager } from '../../utils/authHelpers';

const Navbar: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const selectedKey =
    location.pathname === '/' ? 'home' :
    location.pathname === '/projects' ? 'projects' :
    location.pathname === '/project/create' ? 'projectCreate' :
    null;

  const handleLogout = () => {
    removeToken();
    navigate('/login');
  };

  const guestMenu = (
    <Menu>
      <Menu.Item key="login">
        <Link to="/login">Login</Link>
      </Menu.Item>
      <Menu.Item key="registration">
        <Link to="/registration">Registration</Link>
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
        mode="horizontal"
        theme="dark"
        selectedKeys={selectedKey ? [selectedKey] : []}
        style={{
          flex: 1,
          borderBottom: 'none',
          justifyContent: 'flex-start', 
        }}
      >
        <Menu.Item key="home" icon={<HomeOutlined />}>
          <Link to="/">Home</Link>
        </Menu.Item>
        {isUserLoggedIn() && (
          <Menu.Item key="projects" icon={<SettingOutlined />}>
            <Link to="/projects">Projects</Link>
          </Menu.Item>
        )}
        {isManager() && (
          <Menu.Item key="projectCreate" icon={<SettingOutlined />}>
            <Link to="/project/create">Create Project</Link>
          </Menu.Item>
        )}
      </Menu>

      <div style={{ display: 'flex', alignItems: 'center' }}>
        {isUserLoggedIn() ? (
          <Button
            type="text"
            icon={<LogoutOutlined />}
            style={{ color: '#fff' }}
            onClick={handleLogout}
          >
            Logout
          </Button>
        ) : (
          <Dropdown overlay={guestMenu} placement="bottomRight">
            <Button
              icon={<UserOutlined />}
              style={{ color: 'white', backgroundColor: 'transparent', border: 'none' }}
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
