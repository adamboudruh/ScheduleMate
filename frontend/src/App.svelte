<script>
  import {Addition} from '../wailsjs/go/main/App.js'

  let num1 = 0;
  let num2 = 0;
  let sum = 0;

  function add() {
    Addition(num1, num2).then(result => sum = result)
  }
</script>

<main>
  <h2>Simple Adder</h2>
  <div class="result" id="result">Sum: {sum}</div>
  <div class="input-box" id="input">
    <input autocomplete="off" bind:value={num1} class="input" id="num1" type="number"/>
    <input autocomplete="off" bind:value={num2} class="input" id="num2" type="number"/>
    <button class="btn" on:click={add}>Add</button>
  </div>
</main>

<style>

  /* .result {
    height: 20px;
    line-height: 20px;
    margin: 1.5rem auto;
  }

  .input-box .btn {
    width: 60px;
    height: 30px;
    line-height: 30px;
    border-radius: 3px;
    border: none;
    margin: 0 0 0 20px;
    padding: 0 8px;
    cursor: pointer;
  }

  .input-box .btn:hover {
    background-image: linear-gradient(to top, #cfd9df 0%, #e2ebf0 100%);
    color: #333333;
  }

  .input-box .input {
    border: none;
    border-radius: 3px;
    outline: none;
    height: 30px;
    line-height: 30px;
    padding: 0 10px;
    background-color: rgba(240, 240, 240, 1);
    -webkit-font-smoothing: antialiased;
  }

  .input-box .input:hover {
    border: none;
    background-color: rgba(255, 255, 255, 1);
  }

  .input-box .input:focus {
    border: none;
    background-color: rgba(255, 255, 255, 1);
  } */

</style>
