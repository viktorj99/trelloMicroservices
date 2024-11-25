import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { addMember, getProject, handleDeleteMember } from '../services/projectService';
import { Button, Form, Input, Modal, notification, Select, Table } from 'antd';
import { User } from '../entities/models/User';
import { useState, useEffect } from 'react';
import { Role } from '../entities/models/Role';
import { createTask, getTasksByProjectId, assignMemberToTask, toggleTaskStatus } from '../services/taskService';
import { Task } from '../entities/models/Task';
import { getTokenData } from '../utils/authHelpers';
import { getAllUserMembers } from '../services/userService';
const { Option } = Select;

const SingleProject = () => {
  const { id } = useParams<{ id: string }>();
  console.log('Project ID from URL:', id);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isTaskModalVisible, setIsTaskModalVisible] = useState(false);
  const [selectedUser, setSelectedUser] = useState<string | null>(null);
  const [taskForm] = Form.useForm();
  const [tasks, setTasks] = useState<Task[]>([]);

  const [userRole, setUserRole] = useState<Role | null>(null); 
  const [userId, setUserId] = useState<string | null>(null);


  const { data: project, isLoading, error } = useQuery({
    queryKey: ['project', id],
    queryFn: () => getProject(id!),
  });


  const { data: users, isLoading: usersLoading } = useQuery<User[]>({
    queryKey: ['users'],
    queryFn: getAllUserMembers,
  });


  useEffect(() => {
    if (id) {
      getTasksByProjectId(id)
        .then((tasks) => setTasks(tasks))  
        .catch(console.error);
    }

    const tokenPayload = getTokenData(); 
    setUserRole(tokenPayload.role);
    setUserId(tokenPayload.id);
  }, [id]);

  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (username: string) => handleDeleteMember(username, project?.members || [], id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', id] });
      notification.success({
        message: 'Success',
        description: 'Project updated successfully!',
      });
    },
    onError: (error: unknown) => {
      notification.error({
        message: 'Error',
        description: `Project update failed: ${(error as Error).message}`,
      });
    },
  });

  const addMutation = useMutation({
    mutationFn: (userId: string) => {
      const userToAdd = users?.find((user) => user.id === userId);
      if (!userToAdd) {
        throw new Error('User not found');
      }
      const newMember = {
        id: userToAdd.id,
        username: userToAdd.username,
        role: userToAdd.role,
      };
      return addMember(newMember, project?.members || [], id!);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project', id] });
      notification.success({
        message: 'Success',
        description: 'Member added successfully!',
      });
      setIsModalVisible(false);
    },
    onError: (error: unknown) => {
      notification.error({
        message: 'Error',
        description: `Member addition failed: ${(error as Error).message}`,
      });
    },
  });
  const taskMutation = useMutation({
    mutationFn: createTask,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      notification.success({
        message: 'Success',
        description: 'Task created successfully!',
      });
      taskForm.resetFields();
      setIsTaskModalVisible(false);
    },
    onError: (error: unknown) => {
      notification.error({
        message: 'Error',
        description: `Task creation failed: ${(error as Error).message}`,
      });
    },
  });

  const assignMutation = useMutation({
    mutationFn: ({ taskId, memberId }: { taskId: string; memberId: string }) =>
      assignMemberToTask(taskId, memberId),
    onSuccess: () => {
      console.log('Member assigned successfully');
    },
    onError: (error) => {
      console.error('Error assigning member', error);
    },
  });

  const handleAssignMemberToTask = (taskId: string, memberId: string) => {
    if (taskId && memberId) {
      assignMutation.mutate({ taskId, memberId });
    }
  };

  const toggleStatusMutation = useMutation({
    mutationFn: ({ taskId, memberId }: { taskId: string; memberId: string }) =>
      toggleTaskStatus(taskId, memberId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      notification.success({
        message: 'Success',
        description: 'Task status updated successfully!',
      });
    },
    onError: (error: unknown) => {
      notification.error({
        message: 'Error',
        description: `Task status update failed: ${(error as Error).message}`,
      });
    },
  });

  const handleToggleTaskStatus = (taskId: string) => {
    if (userId) {
      toggleStatusMutation.mutate({ taskId, memberId: userId });
    }
  };

  if (isLoading || usersLoading) {
    return <p>Loading project data...</p>;
  }

  if (error) {
    return <p>Error: {(error as Error).message}</p>;
  }

  if (!project) {
    return <p>No project data available.</p>;
  }

  const handleAddMember = () => {
    if (selectedUser) {
      addMutation.mutate(selectedUser);
    } else {
      notification.error({
        message: 'Error',
        description: 'Please select a user to add.',
      });
    }
  };

  const handleCreateTask = (values: any) => {
    taskMutation.mutate(values);
  };

  return (
    <div>
      <h1>{project.name}</h1>
      <p>Expected End Date: {project.expectedEndDate}</p>
      <p>Max Members: {project.maxMembers}</p>
      <p>Min Members: {project.minMembers}</p>

      {userRole === Role.Manager && (
        <>
          <Button type='primary' onClick={() => setIsModalVisible(true)}>
            Add Member
          </Button>
          <Button type='primary' onClick={() => setIsTaskModalVisible(true)} style={{ marginLeft: 16 }}>
            Create Task
          </Button>
        </>
      )}

      {/* Modal for adding a member */}
      <Modal
        title='Add a Member'
        open={isModalVisible}
        onOk={handleAddMember}
        onCancel={() => setIsModalVisible(false)}
      >
        <Select
          placeholder='Select a user'
          style={{ width: '100%' }}
          onChange={(value) => setSelectedUser(value)}
        >
          {users?.map((user) => (
            <Option key={user.id} value={user.id}>
              {user.username}
            </Option>
          ))}
        </Select>
      </Modal>

      {/* Modal for creating a task */}
      <Modal
        title='Create Task'
        open={isTaskModalVisible}
        onCancel={() => setIsTaskModalVisible(false)}
        footer={null}
      >
        <Form form={taskForm} layout='vertical' onFinish={handleCreateTask}>
          <Form.Item
            label='Task Title'
            name='title'
            rules={[{ required: true, message: 'Please enter the task title!' }]}
          >
            <Input placeholder='Enter task title' />
          </Form.Item>

          <Form.Item
            label='Description'
            name='description'
            rules={[{ required: true, message: 'Please enter the description!' }]}
          >
            <Input.TextArea placeholder='Enter task description' />
          </Form.Item>

          <Form.Item name='status' initialValue='PENDING' hidden>
            <Input type='hidden' />
          </Form.Item>

          <Form.Item label='Project ID' name='project' initialValue={id} hidden />

          <Form.Item>
            <Button type='primary' htmlType='submit'>
              Create Task
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      {/* Tasks Table */}
      <h2>Tasks</h2>
      <Table dataSource={tasks} rowKey='id'>
        <Table.Column title='Task Title' dataIndex='title' />
        <Table.Column title='Description' dataIndex='description' />
        <Table.Column title='Status' dataIndex='status' />
        <Table.Column
          title={userRole === Role.Member ? 'Change Status' : 'Assign Member'}
          render={(_, task: any) => {
            if (userRole === Role.Member && task.member === userId) {
              return (
                <Button
                  onClick={() => handleToggleTaskStatus(task.id)}
                >
                  {task.status === 'IN_PROGRESS' ? 'Mark Finished' : 'Mark In Progress'}
                </Button>
              );
            } else if (userRole === Role.Manager) {
              if (task.status !== 'PENDING') {
                return <span>Assigned</span>;
              } else {
                return (
                  <Select
                  style={{ width: 200 }}
                  onChange={(value) => handleAssignMemberToTask(task.id!, value)}
                >
                  {project.members.map((member: User) => (
                    <Option key={member.id} value={member.id}>
                      {member.username}
                    </Option>
                  ))}
                </Select>
                );
              }
            }
            return <span>Assigned</span>;
          }}
        />
      </Table>

      {/* Members Table */}
      <Table dataSource={project.members} rowKey='username'>
        <Table.Column title='Username' dataIndex='username' />
        <Table.Column title='Role' dataIndex='role' />
        <Table.Column
          title='Action'
          render={(_, record: User) => (
            <Button danger onClick={() => mutation.mutate(record.username)}>
              Delete
            </Button>
          )}
        />
      </Table>
    </div>
  );
};

export default SingleProject;
