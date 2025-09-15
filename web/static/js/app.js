/**
 * Gopher Chat - Discord-like WebSocket Chat Client
 * 
 * This JavaScript file handles:
 * - WebSocket connection management
 * - Room management and switching
 * - Real-time messaging
 * - User interface interactions
 * - Authentication and user management
 */

class GopherChat {
    constructor() {
        this.ws = null;
        this.currentRoom = '';
        this.currentUsername = '';
        this.token = '';
        this.isConnected = false;
        this.rooms = [];
        this.members = [];
        
        // Bind methods
        this.connect = this.connect.bind(this);
        this.handleMessage = this.handleMessage.bind(this);
        this.sendMessage = this.sendMessage.bind(this);
        
        // Initialize the application
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.showLoginModal();
    }

    setupEventListeners() {
        // Login form
        const loginBtn = document.getElementById('loginBtn');
        const usernameInput = document.getElementById('usernameInput');
        
        loginBtn.addEventListener('click', this.handleLogin.bind(this));
        usernameInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.handleLogin();
            }
        });

        // Message input
        const messageInput = document.getElementById('messageInput');
        const sendBtn = document.getElementById('sendBtn');
        
        sendBtn.addEventListener('click', this.sendMessage);
        messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        // Room management
        const refreshRoomsBtn = document.getElementById('refreshRooms');
        refreshRoomsBtn.addEventListener('click', this.refreshRooms.bind(this));

        // User management
        const editUsernameBtn = document.getElementById('editUsernameBtn');
        const leaveRoomBtn = document.getElementById('leaveRoomBtn');
        
        editUsernameBtn.addEventListener('click', this.showUsernameModal.bind(this));
        leaveRoomBtn.addEventListener('click', this.leaveCurrentRoom.bind(this));

        // Username modal
        const saveUsernameBtn = document.getElementById('saveUsernameBtn');
        const cancelUsernameBtn = document.getElementById('cancelUsernameBtn');
        const newUsernameInput = document.getElementById('newUsernameInput');
        
        saveUsernameBtn.addEventListener('click', this.saveUsername.bind(this));
        cancelUsernameBtn.addEventListener('click', this.hideUsernameModal.bind(this));
        newUsernameInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.saveUsername();
            }
        });

        // Members sidebar
        const membersBtn = document.getElementById('membersBtn');
        const closeMembersBtn = document.getElementById('closeMembersBtn');
        
        membersBtn.addEventListener('click', this.toggleMembersSidebar.bind(this));
        closeMembersBtn.addEventListener('click', this.hideMembersSidebar.bind(this));

        // Handle window beforeunload
        window.addEventListener('beforeunload', () => {
            if (this.ws) {
                this.ws.close();
            }
        });
    }

    async handleLogin() {
        const usernameInput = document.getElementById('usernameInput');
        const loginError = document.getElementById('loginError');
        const username = usernameInput.value.trim();

        if (!username) {
            this.showError(loginError, 'Username cannot be empty');
            return;
        }

        if (username.length < 2 || username.length > 20) {
            this.showError(loginError, 'Username must be between 2 and 20 characters');
            return;
        }

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ username }),
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Login failed');
            }

            const data = await response.json();
            this.token = data.token;
            this.currentUsername = username;

            // Update UI
            this.updateUserAvatar(username);
            document.getElementById('currentUsername').textContent = username;
            
            // Hide login modal and show chat
            this.hideLoginModal();
            this.showChatApp();
            
            // Connect WebSocket
            this.connect();

        } catch (error) {
            console.error('Login error:', error);
            this.showError(loginError, error.message);
        }
    }

    connect() {
        if (this.ws) {
            this.ws.close();
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?token=${this.token}`;
        
        this.ws = new WebSocket(wsUrl);
        
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.isConnected = true;
            this.updateConnectionStatus('Connected', true);
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        };

        this.ws.onclose = (event) => {
            console.log('WebSocket closed:', event.code, event.reason);
            this.isConnected = false;
            this.updateConnectionStatus('Disconnected', false);
            
            // Attempt to reconnect after 3 seconds if not intentional
            if (event.code !== 1000) {
                setTimeout(() => {
                    if (!this.isConnected) {
                        console.log('Attempting to reconnect...');
                        this.updateConnectionStatus('Reconnecting...', false);
                        this.connect();
                    }
                }, 3000);
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.updateConnectionStatus('Connection Error', false);
        };
    }

    handleMessage(message) {
        console.log('Received message:', message);
        
        switch (message.type) {
            case 'welcome':
                this.addSystemMessage(message.message);
                break;
            
            case 'room_list':
                this.updateRoomList(message.data);
                break;
            
            case 'room_joined':
                this.handleRoomJoined(message);
                break;
            
            case 'room_left':
                this.handleRoomLeft(message);
                break;
            
            case 'room_message':
                if (message.room === this.currentRoom) {
                    this.addMessage(message.username, message.message, message.timestamp);
                }
                break;
            
            case 'user_joined':
            case 'user_left':
            case 'user_renamed':
                if (message.room === this.currentRoom) {
                    this.addSystemMessage(message.message);
                }
                break;
            
            case 'username_changed':
                this.currentUsername = message.message.split(' ').pop();
                document.getElementById('currentUsername').textContent = this.currentUsername;
                this.updateUserAvatar(this.currentUsername);
                this.addSystemMessage(message.message);
                break;
            
            default:
                console.log('Unknown message type:', message.type);
        }
    }

    sendMessage() {
        const input = document.getElementById('messageInput');
        const message = input.value.trim();

        if (!message || !this.currentRoom || !this.isConnected) {
            return;
        }

        this.sendWebSocketMessage({
            type: 'room_message',
            message: message
        });

        input.value = '';
        input.focus();
    }

    sendWebSocketMessage(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            console.error('WebSocket not connected');
        }
    }

    joinRoom(roomName) {
        if (roomName === this.currentRoom) {
            return;
        }

        this.sendWebSocketMessage({
            type: 'join_room',
            data: roomName
        });
    }

    leaveCurrentRoom() {
        if (!this.currentRoom) {
            return;
        }

        this.sendWebSocketMessage({
            type: 'leave_room'
        });
    }

    refreshRooms() {
        this.sendWebSocketMessage({
            type: 'get_rooms'
        });
    }

    handleRoomJoined(message) {
        this.currentRoom = message.room;
        this.updateCurrentRoomDisplay(message.room);
        this.enableMessageInput();
        this.clearMessages();
        this.addSystemMessage(message.message);
        
        // Update room list to show active room
        this.updateActiveRoom(message.room);
    }

    handleRoomLeft(message) {
        this.currentRoom = '';
        this.updateCurrentRoomDisplay('Select a room');
        this.disableMessageInput();
        this.addSystemMessage(message.message);
        
        // Update room list to remove active state
        this.updateActiveRoom('');
    }

    updateRoomList(rooms) {
        const roomList = document.getElementById('roomList');
        roomList.innerHTML = '';

        if (!rooms || rooms.length === 0) {
            const emptyItem = document.createElement('div');
            emptyItem.className = 'channel-item';
            emptyItem.innerHTML = `
                <span class="channel-icon">#</span>
                <span class="channel-name">No rooms available</span>
            `;
            roomList.appendChild(emptyItem);
            return;
        }

        rooms.forEach(room => {
            const roomItem = document.createElement('div');
            roomItem.className = 'channel-item';
            if (room.name === this.currentRoom) {
                roomItem.classList.add('active');
            }
            
            roomItem.innerHTML = `
                <span class="channel-icon">#</span>
                <span class="channel-name">${this.escapeHtml(room.name)}</span>
                <span class="channel-member-count">${room.memberCount}</span>
            `;
            
            roomItem.addEventListener('click', () => {
                this.joinRoom(room.name);
            });
            
            roomList.appendChild(roomItem);
        });
    }

    updateActiveRoom(roomName) {
        const roomItems = document.querySelectorAll('.channel-item');
        roomItems.forEach(item => {
            const nameElement = item.querySelector('.channel-name');
            if (nameElement && nameElement.textContent === roomName) {
                item.classList.add('active');
            } else {
                item.classList.remove('active');
            }
        });
    }

    updateCurrentRoomDisplay(roomName) {
        const roomNameElement = document.getElementById('currentRoomName');
        const roomDescription = document.getElementById('roomDescription');
        const messageInput = document.getElementById('messageInput');
        
        roomNameElement.textContent = roomName;
        
        if (roomName === 'Select a room') {
            roomDescription.textContent = 'Choose a room from the sidebar to start chatting';
            messageInput.placeholder = 'Select a room to start messaging...';
        } else {
            roomDescription.textContent = `Welcome to #${roomName}`;
            messageInput.placeholder = `Message #${roomName}`;
        }
    }

    addMessage(username, message, timestamp = null) {
        const messagesWrapper = document.getElementById('messagesWrapper');
        const messageElement = document.createElement('div');
        messageElement.className = 'message';
        
        const time = timestamp ? new Date(timestamp) : new Date();
        const timeString = time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        
        const avatarInitial = username.charAt(0).toUpperCase();
        
        messageElement.innerHTML = `
            <div class="message-avatar">
                <div class="avatar-img">${avatarInitial}</div>
            </div>
            <div class="message-content">
                <div class="message-header">
                    <span class="message-username">${this.escapeHtml(username)}</span>
                    <span class="message-timestamp">${timeString}</span>
                </div>
                <div class="message-text">${this.escapeHtml(message)}</div>
            </div>
        `;
        
        messagesWrapper.appendChild(messageElement);
        this.scrollToBottom();
    }

    addSystemMessage(message) {
        const messagesWrapper = document.getElementById('messagesWrapper');
        const messageElement = document.createElement('div');
        messageElement.className = 'system-message';
        messageElement.textContent = message;
        
        messagesWrapper.appendChild(messageElement);
        this.scrollToBottom();
    }

    clearMessages() {
        const messagesWrapper = document.getElementById('messagesWrapper');
        messagesWrapper.innerHTML = '';
    }

    scrollToBottom() {
        const messagesContainer = document.getElementById('messagesContainer');
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    enableMessageInput() {
        const messageInput = document.getElementById('messageInput');
        const sendBtn = document.getElementById('sendBtn');
        
        messageInput.disabled = false;
        sendBtn.disabled = false;
        messageInput.focus();
    }

    disableMessageInput() {
        const messageInput = document.getElementById('messageInput');
        const sendBtn = document.getElementById('sendBtn');
        
        messageInput.disabled = true;
        sendBtn.disabled = true;
    }

    updateConnectionStatus(status, isConnected) {
        const statusText = document.getElementById('connectionStatus').querySelector('.status-text');
        const statusDot = document.getElementById('connectionStatus').querySelector('.status-dot');
        
        statusText.textContent = status;
        
        if (isConnected) {
            statusDot.classList.add('connected');
        } else {
            statusDot.classList.remove('connected');
        }
    }

    updateUserAvatar(username) {
        const userAvatar = document.getElementById('userAvatar');
        userAvatar.textContent = username.charAt(0).toUpperCase();
    }

    // Username management
    showUsernameModal() {
        const modal = document.getElementById('usernameModal');
        const input = document.getElementById('newUsernameInput');
        modal.style.display = 'flex';
        input.value = this.currentUsername;
        input.focus();
        input.select();
    }

    hideUsernameModal() {
        const modal = document.getElementById('usernameModal');
        const error = document.getElementById('usernameError');
        modal.style.display = 'none';
        error.textContent = '';
    }

    async saveUsername() {
        const input = document.getElementById('newUsernameInput');
        const error = document.getElementById('usernameError');
        const newUsername = input.value.trim();

        if (!newUsername) {
            this.showError(error, 'Username cannot be empty');
            return;
        }

        if (newUsername.length < 2 || newUsername.length > 20) {
            this.showError(error, 'Username must be between 2 and 20 characters');
            return;
        }

        if (newUsername === this.currentUsername) {
            this.hideUsernameModal();
            return;
        }

        try {
            // Use WebSocket method for immediate response
            this.sendWebSocketMessage({
                type: 'change_username',
                data: newUsername
            });
            
            this.hideUsernameModal();
            
        } catch (error) {
            console.error('Username change error:', error);
            this.showError(error, 'Failed to change username');
        }
    }

    // Members sidebar management
    toggleMembersSidebar() {
        const sidebar = document.getElementById('membersSidebar');
        const isVisible = sidebar.style.display !== 'none';
        
        if (isVisible) {
            this.hideMembersSidebar();
        } else {
            this.showMembersSidebar();
        }
    }

    showMembersSidebar() {
        const sidebar = document.getElementById('membersSidebar');
        sidebar.style.display = 'flex';
        // In a real implementation, you'd fetch and display current room members
    }

    hideMembersSidebar() {
        const sidebar = document.getElementById('membersSidebar');
        sidebar.style.display = 'none';
    }

    // UI helper methods
    showLoginModal() {
        const modal = document.getElementById('loginModal');
        modal.style.display = 'flex';
        document.getElementById('usernameInput').focus();
    }

    hideLoginModal() {
        const modal = document.getElementById('loginModal');
        modal.style.display = 'none';
    }

    showChatApp() {
        const chatApp = document.getElementById('chatApp');
        chatApp.style.display = 'flex';
    }

    showError(errorElement, message) {
        errorElement.textContent = message;
        setTimeout(() => {
            errorElement.textContent = '';
        }, 5000);
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize the chat application when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new GopherChat();
});

// Export for potential use in other scripts
window.GopherChat = GopherChat;