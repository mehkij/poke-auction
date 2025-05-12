import Button from "./components/Button";
import Status from "./components/Status";
import GHLogo from "./assets/github-mark.svg";
import DiscordLogo from "./assets/Discord-Symbol-White.svg";
import PokeBall from "./assets/Poke_Ball.webp";

function App() {
  return (
    <div className="h-screen w-screen flex flex-col justify-center items-center">
      <img className="size-64 m-4" src={PokeBall}></img>
      <Status online={false} />
      <div className="flex m-4 gap-4">
        <Button
          className="flex gap-2 items-center justify-center bg-red-500 hover:bg-red-600 text-white border-2 border-red-600"
          // onClick={() => window.open("", "_blank")}
        >
          <img className="size-4" src={DiscordLogo}></img>
          <p>Coming soon...</p>
        </Button>

        <Button
          className="flex gap-2 items-center justify-center bg-white hover:bg-gray-100 border-2 border-gray-100"
          onClick={() =>
            window.open("https://github.com/mehkij/poke-auction", "_blank")
          }
        >
          <img className="size-4" src={GHLogo}></img>
          <p>GitHub</p>
        </Button>
      </div>
    </div>
  );
}

export default App;
