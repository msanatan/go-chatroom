var app = new Vue({
    el: '#app',
    data: {
        ws: null,
        serverUrl: "ws://localhost:8080/api/ws",
        messages: [],
        rooms: [],
        newMessage: "",
        loginDetails: {
            username: "",
            password: "",
        },
        registrationDetails: {
            email: "",
            username: "",
            password: "",
        },
        user: {
            username: "",
            token: "",
        },
        authError: "",
        registerSuccess: "",
        room: {
            id: 0,
            name: ""
        },
        newRoom: "",
        loggedIn: false,
        inChat: false,
    },
    mounted() {
        if (!this.loggedIn) {
            if (localStorage.token) {
                this.user.token = localStorage.token;
                this.loggedIn = true;
                this.getRooms();
            }
        }
    },
    methods: {
        async login() {
            try {
                const response = await axios.post(`http://${location.host}/login`, this.loginDetails);
                this.user.username = this.loginDetails.username;
                this.user.token = response.data.token;
                this.loggedIn = true;
                localStorage.token = this.user.token;
                this.getRooms();
            } catch (e) {
                this.authError = e.response.data.error;
                console.error(e);
                console.error(this.authError);
            }
        },
        async register() {
            try {
                const response = await axios.post(`http://${location.host}/register`, this.registrationDetails);
                this.registerSuccess = "Successfully registered! Please log in";
            } catch (e) {
                this.authError = e.response.data.error;
                console.error(e);
                console.error(this.authError);
            }
        },
        async sendMessage() {
            if (this.newMessage !== "") {
                try {
                    const response = await axios.post(`http://${location.host}/api/messages`,
                        {
                            message: this.newMessage,
                            type: "user",
                            username: this.user.username,
                            roomId: this.room.id,
                        }, {
                        headers: {
                            'Content-Type': 'application/json',
                            'Authorization': 'Bearer ' + this.user.token
                        }
                    });

                    this.newMessage = "";
                } catch (e) {
                    console.error(e);
                }
            }
        },
        async getLatestMessages() {
            try {
                const response = await axios.get(`http://${location.host}/api/rooms/${this.room.id}/messages`, {
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer ' + this.user.token
                    }
                });

                return response.data;
            } catch (e) {
                console.error(e);
            }
        },
        async createRoom() {
            if (this.newRoom !== "") {
                try {
                    const response = await axios.post(`http://${location.host}/api/rooms`,
                        {
                            name: this.newRoom,
                        },
                        {
                            headers: {
                                'Content-Type': 'application/json',
                                'Authorization': 'Bearer ' + this.user.token
                            }
                        });
                    this.rooms.push(response.data);
                    this.newRoom = "";
                } catch (e) {
                    console.error(e);
                }
            }
        },
        async getRooms() {
            try {
                const response = await axios.get(`http://${location.host}/api/rooms`, {
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer ' + this.user.token
                    }
                });

                this.rooms = response.data.rooms || [];
            } catch (e) {
                console.error(e);
            }
        },
        async connectToWebsocket() {
            if (this.user.token !== "") {
                try {
                    // First populate chat with the most recent messages
                    const lastMessages = await this.getLatestMessages();
                    this.messages = lastMessages.messages ? lastMessages.messages : [];
                    console.log('Retrieved latest messages');

                    // Then connect to the websocket server
                    this.ws = new WebSocket(`${this.serverUrl}/${this.room.id}?bearer=${this.user.token}`);
                    this.ws.addEventListener('open', (event) => { this.onWebsocketOpen(event) });
                    this.ws.addEventListener('message', (event) => { console.log(event); this.handleNewMessage(event) });
                } catch (e) {
                    console.error(e);
                }
            }
        },
        onWebsocketOpen() {
            console.log("Connected to chat room");
        },
        handleNewMessage(event) {
            let msg = JSON.parse(event.data);
            if (this.messages.length === 50) {
                this.messages.pop();
            }
            this.messages.push(msg);
        },
        handleSelectRoom(room) {
            this.room = room;
            this.inChat = true;
            this.connectToWebsocket();
        }
    }
});