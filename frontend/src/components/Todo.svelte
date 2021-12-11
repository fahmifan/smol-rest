<script lang="ts">
  import { Client, SmolService, AddTodoRequest, FindAllTodosFilter, Todo } from "../service/oto.gen";
  import { onMount, } from 'svelte'
  const service = new SmolService(new Client())
  let todos:  Todo[] = []
  let todoItem = ''
  const addTodo = async (detail: string, done: boolean) => {
    const req = new AddTodoRequest()
    req.detail = detail
    req.done = done

    try {
     const res = await service.addTodo(req)
     todos.push(res)
     todos = todos
     todoItem = ''
    } catch (err) {
      console.error(err)
    }
  }

  onMount(async () => {
    const filter = new FindAllTodosFilter()
    filter.page = 1
    filter.size = 10
    const res = await service.findAllTodos(filter)
    todos.push(...res.todos)
    todos = todos
  })

  const handleAddTodo = () => {
    addTodo(todoItem, false)
  }
</script>

<main>
  <br>
  <br>

  <form on:submit|preventDefault={handleAddTodo}>
    <input type="text" bind:value={todoItem}>
  </form>

  <ul>
    {#each todos as todo}
    <li id={todo.id}>
      <p>{todo.detail}</p>
    </li>
    {/each}
  </ul>
</main>

<style>
  button {
    font-family: inherit;
    font-size: inherit;
    padding: 1em 2em;
    color: #ff3e00;
    background-color: rgba(255, 62, 0, 0.1);
    border-radius: 2em;
    border: 2px solid rgba(255, 62, 0, 0);
    outline: none;
    width: 200px;
    font-variant-numeric: tabular-nums;
    cursor: pointer;
  }
  button:focus {
    border: 2px solid #ff3e00;
  }
  button:active {
    background-color: rgba(255, 62, 0, 0.2);
  }
</style>
