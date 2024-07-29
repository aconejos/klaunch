import subprocess

def install_with_homebrew(package_name):
    try:
        # Attempt to install the package using Homebrew
        subprocess.check_call(['brew', 'list', package_name])
        print(f"{package_name} is already installed via Homebrew.")
    except subprocess.CalledProcessError:
        # If the package is not listed, attempt to install it
        print(f"Installing {package_name} via Homebrew...")
        subprocess.check_call(['brew', 'install', package_name])

def main():
    # Ensure Homebrew is installed
    try:
        subprocess.check_call(['brew', '--version'])
        print("Homebrew is installed.")
    except subprocess.CalledProcessError:
        print("Homebrew is not installed. Please install Homebrew first.")
        return
    
    # Install Python 3 if not already installed
    install_with_homebrew('python@3')

    # Install requests and python-dotenv if they are not already installed
    install_with_homebrew('python@3')
    install_with_homebrew('pyenv')

if __name__ == "__main__":
    main()
