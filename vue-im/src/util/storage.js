const saveData = function (key, data) {
  let localData = window.localStorage.getItem('VueIM')
  if (!localData) {
    localData = {
      [key]: data
    }
  } else {
    localData = JSON.parse(localData)
    localData[key] = data
  }
  window.localStorage.setItem('VueIM', JSON.stringify(localData))
}

const readData = function (key) {
  const localData = window.localStorage.getItem('VueIM')
  return localData ? JSON.parse(localData)[key] : localData
}

const getItem = function (name) {
  return window.localStorage.getItem(name)
}

const setItem = function (name, data) {
  return window.localStorage.setItem(name, data)
}

export default {
  saveData,
  readData,
  getItem,
  setItem
}
