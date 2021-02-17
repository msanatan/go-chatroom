var app = new Vue({
    el: '#app',
    data: {
        ws: null,
        serverUrl: "ws://localhost:8080/api/ws",
        messages: [],
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
        currentRoom: null,
    },
    mounted() {
        if (localStorage.token) {
            this.user.token = localStorage.token;
            this.connectToWebsocket();
        }
    },
    methods: {
        async login() {
            try {
                const response = await axios.post("http://" + location.host + '/login', this.loginDetails);
                this.user.username = this.loginDetails.username;
                this.user.token = response.data.token;
                localStorage.token = this.user.token;
                this.connectToWebsocket();
            } catch (e) {
                this.authError = e.response.data.error;
                console.error(e);
                console.error(this.authError);
            }
        },
        async register() {
            try {
                const response = await axios.post("http://" + location.host + '/register', this.registrationDetails);
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
                    const response = await axios.post("http://" + location.host + '/api/messages',
                        {
                            message: this.newMessage,
                            type: "user",
                            username: this.user.username,
                            roomId: this.currentRoom,
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
                const response = await axios.get(`http://${location.host}/api/rooms/${this.currentRoom}/messages`, {
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
        async connectToWebsocket() {
            if (this.user.token !== "") {
                try {
                    // First populate chat with the most recent messages
                    const lastMessages = await this.getLatestMessages();
                    this.messages = lastMessages.messages ? lastMessages.messages.reverse() : [];
                    console.log('Retrieved latest messages');

                    // Then connect to the websocket server
                    this.ws = new WebSocket(`${this.serverUrl}/${this.currentRoom}?bearer=${this.user.token}`);
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
            this.messages.push(msg);
        },
    }
});