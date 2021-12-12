<script lang="ts">
  import Todo from "@/components/Todo.svelte";
  import { onMount } from "svelte";
  import { User, Empty, SmolService, Client } from "../../service/oto.gen"
  
  const service = new SmolService(new Client())
  let user: User

  const logout = async () => {
    try {
      await service.logoutUser(new Empty())
      console.log('success logout user')
    } catch (err) {
      console.error(err)
    }
  }

  const findUser = async () => {
    try {
      const res = await service.findCurrentUser(new Empty())
      if (!res) {
        return
      }
      user = res
    } catch (err) {
      console.error(err)
    }
  }

  onMount(() => {
    findUser()
  })

</script>

<main>
  <p>Hello {user && user.email}</p>
  <button on:click={logout}>Logout</button>

  <h1>Add Todo</h1>

  <Todo />
</main>
<style>
  :root {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen,
      Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
  }
  main {
    text-align: center;
    padding: 1em;
    margin: 0 auto;
  }
  img {
    height: 16rem;
    width: 16rem;
  }
  h1 {
    color: #ff3e00;
    text-transform: uppercase;
    font-size: 4rem;
    font-weight: 100;
    line-height: 1.1;
    margin: 2rem auto;
    max-width: 14rem;
  }
  p {
    max-width: 14rem;
    margin: 1rem auto;
    line-height: 1.35;
  }
  @media (min-width: 480px) {
    h1 {
      max-width: none;
    }
    p {
      max-width: none;
    }
  }
</style>
