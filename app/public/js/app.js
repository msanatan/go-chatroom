var app = new Vue({
    el: '#app',
    data: {
        ws: null,
        serverUrl: "ws://localhost:8080/api/ws",
        messages: [],
        newMessage: "",
        user: {
            username: "",
            password: "",
            token: ""
        },
        authError: "",
        registerSuccess: ""
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
                const response = await axios.post("http://" + location.host + '/login', this.user);
                this.user.token = response.data.token;
                localStorage.token = this.user.token;
                this.connectToWebsocket();
            } catch (e) {
                this.authError = e.response.data.error;
                console.log(e);
                console.log(this.authError);
            }
        },
        async register() {
            try {
                const response = await axios.post("http://" + location.host + '/register', this.user);
                this.registerSuccess = "Successfully registered! Please log in";
            } catch (e) {
                this.authError = e.response.data.error;
                console.log(e);
                console.log(this.authError);
            }
        },
        connectToWebsocket() {
            if (this.user.token !== "") {
                this.ws = new WebSocket(this.serverUrl + "?bearer=" + this.user.token);
                this.ws.addEventListener('open', (event) => { this.onWebsocketOpen(event) });
                this.ws.addEventListener('message', (event) => { console.log(event); this.handleNewMessage(event) });
            }
        },
        onWebsocketOpen() {
            console.log("Connected to chat room");
        },
        handleNewMessage(event) {
            let data = event.data;
            data = data.split(/\r?\n/);

            for (let i = 0; i < data.length; i++) {
                let msg = JSON.parse(data[i]);
                this.messages.push(msg);
            }
        },
        sendMessage() {
            if (this.newMessage !== "") {
                this.ws.send(JSON.stringify({ message: this.newMessage }));
                this.newMessage = "";
            }
        }

    }
});