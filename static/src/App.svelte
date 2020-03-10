<script>
  import { onMount } from "svelte";

  let currencySign = "£";
  let balance = 0.0;
  let moneyPlus = 0.0;
  let moneyMinus = 0.0;
  let err = "";
  let text = "";
  let amount = "";
  let transactions = [];

  const getTransactions = () => {
    fetch("transaction")
      .then(res => res.json())
      .then(res => {
        if (res.data.length > 0) transactions = res.data;
        updateValues();
      });
  };

  const updateValues = () => {
    const amounts = transactions.map(t => t.amount);
    balance = amounts.reduce((acc, item) => (acc += item), 0).toFixed(2);
    moneyPlus = amounts
      .filter(item => item > 0)
      .reduce((acc, item) => (acc += item), 0)
      .toFixed(2);

    moneyMinus = (
      amounts.filter(item => item < 0).reduce((acc, item) => (acc += item), 0) *
      -1
    ).toFixed(2);
  };

  const submitForm = () => {
    if (!text || !amount) {
      err = "Please fill in both fields";
    } else {
      err = "";
      let newTransaction = {
        text,
        amount
      };
      fetch("transaction", {
        method: "POST",
        body: JSON.stringify(newTransaction)
      })
        .then(res => res.json())
        .then(res => {
          transactions = [res.transaction, ...transactions];
          updateValues();
          text = "";
          amount = "";
        });
    }
  };

  const deleteTransaction = id => {
    fetch(`transaction/${id}`, {
      method: "DELETE"
    })
      .then(res => res.json())
      .then(res => {
        transactions = transactions.filter(t => t.id !== id);
        updateValues();
      });
  };

  const generateID = () => Math.floor(Math.random() * 100000000);

  onMount(() => getTransactions());
</script>

<h2>Expense Tracker</h2>

<div class="container">
  <h4 class="head-stuff">Your Balance</h4>
  <h1 class="head-stuff" id="balance">{currencySign}{balance}</h1>

  <div class="inc-exp-container">
    <div>
      <h4>Income</h4>
      <p id="money-plus" class="money plus">+{currencySign}{moneyPlus}</p>
    </div>
    <div>
      <h4>Expense</h4>
      <p id="money-minus" class="money minus">-{currencySign}{moneyMinus}</p>
    </div>
  </div>

  <h3>History</h3>
  {#if transactions.length > 0}
    <ul id="list" class="list">
      {#each transactions as transaction}
        <li class={transaction.amount < 0 ? 'minus' : 'plus'}>
          {transaction.text}
          <span>
            {transaction.amount < 0 ? '-' : '+'}{currencySign}{Math.abs(transaction.amount)}
          </span>
          <button
            class="delete-btn"
            on:click={() => deleteTransaction(transaction.id)}>
            x
          </button>
        </li>
      {/each}
    </ul>
  {:else}
    <h4>Add transaction</h4>
  {/if}

  <h3>Add new transaction</h3>

  <form id="form">
    <div class="errors">{err}</div>
    <div class="form-control">
      <label for="text">Text</label>
      <input
        od="text"
        placeholder="Enter text..."
        type="text"
        bind:value={text} />
    </div>
    <div class="form-control">
      <label for="amount">
        Amount
        <br />
        (negative - expense, positive - income)
      </label>
      <input
        id="amount"
        placeholder="Enter amount..."
        type="number"
        bind:value={amount} />
    </div>
    <button class="btn" on:click|preventDefault={submitForm}>
      Add transaction
    </button>
  </form>
</div>
