<script setup lang="ts">
import { ref } from 'vue';
import { Greet } from "../../wailsjs/go/main/App";

const name = ref("hi")
const response = ref("response...")

function getGreet() {
  Greet(name.value).then((res) => response.value = res);
}
</script>

<template>
  <div class="home">
    hello HOME
    <input type="text" v-model="name">
    <button @click="getGreet">greet me</button>
    <div>{{ response }}</div>
  </div>
</template>

<style lang="scss">
.home {
  button {
    display: inline-block;
    min-width: 80px;
    height: 30px;
    line-height: 30px;
    padding: 0 5px;
    margin-left: 8px;
    background-color: #3459b0;
    border-radius: 2px;
    text-align: center;
    text-decoration: none;
    color: #000000;
    font-size: 14px;

    &:hover {
      background-color: #4879eb;
      color: #ffffff;
      cursor: pointer;
    }
  }
}
</style>
