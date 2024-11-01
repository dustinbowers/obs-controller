<script setup lang="ts">
import { ref } from 'vue'
import { useLogStore } from "@/stores/logs"

const logStore = useLogStore();

</script>

<template>
    <section>
        <textarea>{{ logStore.history }}
        </textarea>
    </section>
</template>

<style lang="scss">
    section {
        display: flex;
        flex-flow: column;
        height: 100%;
        width: 100%;
        padding: 16px 16px;

        textarea {
            padding: 8px; 
            flex: 1 1 auto;
            font-family: "Courier New";
            font-size: 14px;
            font-weight: 500; 
            white-space: pre;
            overflow-wrap: normal;
            overflow-x: scroll;
        }
    }
</style>
