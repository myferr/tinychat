<template>
    <div class="app-container" style="display: flex; height: 100vh">
        <Sidebar
            class="sidebar"
            :models="models"
            :selectedModel="selectedModel"
            @selectModel="selectModel"
        />

        <main
            class="chat-container"
            style="flex-grow: 1; display: flex; flex-direction: column"
        >
            <section
                class="messages"
                style="
                    flex-grow: 1;
                    padding: 1rem;
                    padding-top: 2rem;
                    overflow-y: auto;
                "
            >
                <div
                    v-for="(msg, idx) in messages"
                    :key="idx"
                    :class="['message', msg.role]"
                    style="margin-bottom: 0.5rem"
                >
                    <strong
                        >{{
                            msg.role === "user" ? "You" : selectedModel
                        }}:</strong
                    >
                    <p>{{ msg.content }}</p>
                </div>
            </section>

            <form
                @submit.prevent="sendMessage"
                style="padding: 1rem; display: flex"
            >
                <input
                    v-model="input"
                    placeholder="Type a message..."
                    style="
                        border: none;
                        outline: none;
                        flex-grow: 1;
                        padding: 1rem;
                        border-radius: 15px;
                        font-size: 1rem;
                        background: #222;
                    "
                    :disabled="!selectedModel"
                />
            </form>
        </main>
    </div>
</template>

<script setup>
import { ref, reactive, watch, onMounted } from "vue";
import Sidebar from "./components/Sidebar.vue";

const models = ref([]);
const selectedModel = ref(null);
const input = ref("");
const messages = reactive([]);

// Load models from backend API
async function loadModels() {
    try {
        const res = await fetch("http://localhost:1111/models");
        if (!res.ok) throw new Error("Failed to fetch models");
        models.value = await res.json();
    } catch (e) {
        console.error(e);
    }
}

// Save/load chat history per model in localStorage
function saveHistory() {
    if (!selectedModel.value) return;
    localStorage.setItem(
        `chat_history_${selectedModel.value}`,
        JSON.stringify(messages),
    );
}

function loadHistory(model) {
    const history = localStorage.getItem(`chat_history_${model}`);
    if (history) {
        messages.splice(0, messages.length, ...JSON.parse(history));
    } else {
        messages.splice(0);
    }
}

function selectModel(model) {
    selectedModel.value = model;
    loadHistory(model);
    input.value = "";
}

// Send message to backend streaming API
async function sendMessage() {
    if (!input.value.trim() || !selectedModel.value) return;
    const userMessage = input.value.trim();
    messages.push({ role: "user", content: userMessage });
    input.value = "";

    // Call backend /generate with model and userMessage
    try {
        const response = await fetch(
            `http://localhost:1111/generate?model=${selectedModel.value}`,
            {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ content: userMessage, max_tokens: 256 }),
            },
        );

        if (!response.ok) throw new Error("Backend error");

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let aiResponse = "";

        // Append streaming text to AI message
        messages.push({ role: "assistant", content: "" });

        while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            const chunk = decoder.decode(value);
            const lines = chunk
                .split("\n")
                .filter((line) => line.startsWith("data: "));

            for (const line of lines) {
                const json = JSON.parse(line.slice(6));
                if (json.response) {
                    aiResponse += json.response;
                    messages[messages.length - 1].content = aiResponse;
                }
            }
        }
    } catch (err) {
        console.error(err);
        messages.push({
            role: "assistant",
            content: "Error: failed to get response.",
        });
    } finally {
        saveHistory();
    }
}

onMounted(() => {
    loadModels();
});
</script>
<style scoped>
.message {
    display: flex;
    flex-direction: column;
}
.message.user p,
.message.assistant p {
    word-wrap: break-word;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
    background-color: #222;
    padding: 0.5rem;
    border-radius: 8px;
    max-width: 90%;
}
</style>
