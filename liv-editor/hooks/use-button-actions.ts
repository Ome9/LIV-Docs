"use client"

import { useState, useCallback } from "react"

export interface ButtonAction {
  id: string
  label: string
  action: () => void | Promise<void>
  isLoading?: boolean
  isDisabled?: boolean
}

export function useButtonActions() {
  const [loadingStates, setLoadingStates] = useState<Record<string, boolean>>({})
  const [notifications, setNotifications] = useState<Array<{ id: string; message: string; type: "success" | "error" }>>(
    [],
  )

  const executeAction = useCallback(async (action: () => void | Promise<void>, buttonId: string) => {
    setLoadingStates((prev) => ({ ...prev, [buttonId]: true }))
    try {
      await action()
      setNotifications((prev) => [
        ...prev,
        { id: Date.now().toString(), message: "Action completed!", type: "success" },
      ])
    } catch (error) {
      setNotifications((prev) => [...prev, { id: Date.now().toString(), message: "Action failed!", type: "error" }])
    } finally {
      setLoadingStates((prev) => ({ ...prev, [buttonId]: false }))
    }
  }, [])

  const clearNotification = useCallback((id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id))
  }, [])

  return {
    loadingStates,
    notifications,
    executeAction,
    clearNotification,
  }
}
