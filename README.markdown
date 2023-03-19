# meteo

## 0. Set up dependencies

Ensure you have asdf installed, then run:

```
ASDF_PYAPP_INCLUDE_DEPS=1 asdf plugin add ansible https://github.com/amrox/asdf-pyapp.git
asdf install
```

## 1. Set up the target hosts

Set up:

 - A user matching your current user, in the `sudo` group.
   ```
   useradd auxesis
   usermod -G sudo -a auxesis
   ```
 - An SSH public key copied over to the Pi
 - Password-less sudo in `/etc/sudoers`:
   ```
   %sudo ALL=(ALL:ALL) NOPASSWD: ALL
   ```

## 2. Set up Ansible hosts

```
echo meteo.example > hosts
```

## 3. Run the bootstrap playbook

```
make
```
