type ButtonProps = {
  className?: string;
} & React.ButtonHTMLAttributes<HTMLButtonElement>;

function Button({ className = "", children, ...props }: ButtonProps) {
  return (
    <button
      className={`px-4 py-2 cursor-pointer rounded shadow-md transition-colors ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}

export default Button;
