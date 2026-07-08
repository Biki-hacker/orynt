import React, { useState, useEffect, useRef, useCallback } from 'react'
import { Task, MedicalRequest } from '../../types'
import { useAuthStore } from '../../store'
import { api } from '../../services/api'
import { wsService } from '../../services/websocket'
import { Badge, Modal, CustomSelect } from '../../components/ui'
import { KanbanSquare, AlertOctagon, HeartHandshake, ShieldAlert, Plus, Check } from 'lucide-react'

const staffOptions = [
  { value: 'volunteer1', label: 'Volunteer 1 (Visitor Services)' },
  { value: 'security1', label: 'Security Guard 1 (Safety)' },
  { value: 'medical1', label: 'First Aider 1 (First Aid)' },
  { value: 'cleaning1', label: 'Cleaner 1 (Facilities)' },
  { value: 'ops1', label: 'Operations Coordinator' }
]

const deptOptions = [
  { value: 'volunteer', label: 'Volunteer' },
  { value: 'security', label: 'Security' },
  { value: 'medical', label: 'Medical' },
  { value: 'cleaning', label: 'Cleaning' },
  { value: 'ops', label: 'Ops' }
]

const priorityOptions = [
  { value: 'low', label: 'Low' },
  { value: 'medium', label: 'Medium' },
  { value: 'high', label: 'High' }
]

const alertTypeOptions = [
  { value: 'crowd', label: 'Crowd Congestion' },
  { value: 'emergency', label: 'First Aid Dispatch' },
  { value: 'weather', label: 'Extreme Weather' },
  { value: 'transport', label: 'Transit Delay' }
]

const alertSeverityOptions = [
  { value: 'info', label: 'Info' },
  { value: 'high', label: 'High Warning' },
  { value: 'critical', label: 'Critical Emergency' }
]

export const StaffDashboard: React.FC = () => {
  const { user } = useAuthStore()
  const [tasks, setTasks] = useState<Task[]>([])
  const [medicalReqs, setMedicalReqs] = useState<MedicalRequest[]>([])
  const [isTaskModalOpen, setIsTaskModalOpen] = useState(false)
  const [isAlertModalOpen, setIsAlertModalOpen] = useState(false)

  // Form states
  const [taskTitle, setTaskTitle] = useState('')
  const [taskDesc, setTaskDesc] = useState('')
  const [taskDept, setTaskDept] = useState('volunteer')
  const [taskAssigned, setTaskAssigned] = useState('volunteer1')
  const [taskPriority, setTaskPriority] = useState('medium')

  const [alertTitle, setAlertTitle] = useState('')
  const [alertContent, setAlertContent] = useState('')
  const [alertType, setAlertType] = useState('crowd')
  const [alertSeverity, setAlertSeverity] = useState('high')

  const loadData = async () => {
    try {
      const [taskData, medData] = await Promise.all([
        api.get<Task[]>('/tasks'),
        api.get<MedicalRequest[]>('/medical')
      ])
      setTasks(taskData)
      setMedicalReqs(medData)
    } catch (err) {
      console.error(err)
    }
  }

  useEffect(() => {
    loadData()

    const unsubTask = wsService.subscribe('task_update', (t: Task) => {
      setTasks((prev) => {
        const exists = prev.some((item) => item.id === t.id)
        if (exists) return prev.map((item) => (item.id === t.id ? t : item))
        return [t, ...prev]
      })
    })

    const unsubMed = wsService.subscribe('medical_request', (req: MedicalRequest) => {
      setMedicalReqs((prev) => [req, ...prev])
    })

    const unsubMedUpdate = wsService.subscribe('medical_update', (req: MedicalRequest) => {
      setMedicalReqs((prev) => prev.map((item) => (item.id === req.id ? req : item)))
    })

    return () => {
      unsubTask()
      unsubMed()
      unsubMedUpdate()
    }
  }, [])

  // ── Drag & Drop state ─────────────────────────────────────────────────────
  const draggingId     = useRef<string | null>(null)
  const [dragOverCol, setDragOverCol] = useState<string | null>(null)
  const ghostRef       = useRef<HTMLDivElement | null>(null)

  const moveTaskToCol = useCallback(async (id: string, newStatus: string) => {
    const task = tasks.find(t => t.id === id)
    if (!task || task.status === newStatus) return
    const validStatus = newStatus as Task['status']
    setTasks(prev => prev.map(t => t.id === id ? { ...t, status: validStatus } : t))
    try {
      await api.put(`/tasks/${id}`, { status: newStatus })
    } catch (err) {
      console.error(err)
      loadData()
    }
  }, [tasks])

  // Mouse / HTML5 DnD
  const handleDragStart = (e: React.DragEvent, id: string) => {
    draggingId.current = id
    e.dataTransfer.effectAllowed = 'move';
    (e.currentTarget as HTMLElement).style.opacity = '0.45'
  }

  const handleDragEnd = (e: React.DragEvent) => {
    (e.currentTarget as HTMLElement).style.opacity = '1'
    draggingId.current = null
    setDragOverCol(null)
  }

  const handleDragOver = (e: React.DragEvent, col: string) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    setDragOverCol(col)
  }

  const handleDragLeave = () => setDragOverCol(null)

  const handleDrop = (e: React.DragEvent, col: string) => {
    e.preventDefault()
    setDragOverCol(null)
    if (draggingId.current) moveTaskToCol(draggingId.current, col)
    draggingId.current = null
  }

  // Touch DnD
  const createGhost = (sourceEl: HTMLElement) => {
    const ghost = sourceEl.cloneNode(true) as HTMLDivElement
    ghost.style.cssText = `
      position:fixed;pointer-events:none;z-index:9999;
      width:${sourceEl.offsetWidth}px;opacity:0.88;
      transform:scale(1.05) rotate(1.5deg);
      border-radius:10px;box-shadow:0 16px 40px rgba(0,0,0,0.2);
    `
    document.body.appendChild(ghost)
    ghostRef.current = ghost
  }

  const moveGhost = (x: number, y: number) => {
    if (!ghostRef.current) return
    ghostRef.current.style.left = `${x - ghostRef.current.offsetWidth / 2}px`
    ghostRef.current.style.top  = `${y - ghostRef.current.offsetHeight / 2}px`
  }

  const removeGhost = () => {
    ghostRef.current?.remove()
    ghostRef.current = null
  }

  const getColFromPoint = (x: number, y: number): string | null => {
    for (const col of ['todo', 'in_progress', 'done']) {
      const el = document.getElementById(`kanban-col-${col}`)
      if (!el) continue
      const r = el.getBoundingClientRect()
      if (x >= r.left && x <= r.right && y >= r.top && y <= r.bottom) return col
    }
    return null
  }

  const handleTouchStart = (e: React.TouchEvent, id: string) => {
    draggingId.current = id
    createGhost(e.currentTarget as HTMLElement)
    const t = e.touches[0]
    moveGhost(t.clientX, t.clientY)
  }

  const handleTouchMove = (e: React.TouchEvent) => {
    e.preventDefault()
    const t = e.touches[0]
    moveGhost(t.clientX, t.clientY)
    setDragOverCol(getColFromPoint(t.clientX, t.clientY))
  }

  const handleTouchEnd = (e: React.TouchEvent) => {
    const t = e.changedTouches[0]
    const col = getColFromPoint(t.clientX, t.clientY)
    removeGhost()
    setDragOverCol(null)
    if (col && draggingId.current) moveTaskToCol(draggingId.current, col)
    draggingId.current = null
  }

  const getColTasks = (status: string) => tasks.filter((t) => t.status === status)

  const colMeta: Record<string, { label: string; border: string }> = {
    todo:        { label: 'TO DO',       border: 'border-zinc-300' },
    in_progress: { label: 'IN PROGRESS', border: 'border-brand-400' },
    done:        { label: 'DONE',        border: 'border-emerald-400' },
  }

  // Form handlers
  const handleCreateTask = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.post('/tasks', {
        title: taskTitle,
        description: taskDesc,
        assignedTo: taskAssigned,
        priority: taskPriority,
        department: taskDept
      })
      setIsTaskModalOpen(false)
      setTaskTitle('')
      setTaskDesc('')
    } catch (err) {
      console.error(err)
    }
  }

  const handleCreateAlert = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.post('/alerts', {
        title: alertTitle,
        content: alertContent,
        type: alertType,
        severity: alertSeverity
      })
      setIsAlertModalOpen(false)
      setAlertTitle('')
      setAlertContent('')
    } catch (err) {
      console.error(err)
    }
  }

  const handleMedicalAction = async (id: string, action: 'assign' | 'resolve') => {
    try {
      if (action === 'assign') {
        await api.put(`/medical/${id}/assign`, { assignedTo: user?.username || 'medical1' })
      } else {
        await api.put(`/medical/${id}/resolve`)
      }
      loadData()
    } catch (err) {
      console.error(err)
    }
  }

  return (
    <div className="space-y-8 font-sans">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-2xl font-extrabold text-neutral-800">Operational Command Board</h1>
          <p className="text-xs text-zinc-500 mt-0.5">Manage live assignments, resolve first aid incidents, and broadcast emergency alerts.</p>
        </div>
        <div className="flex gap-2 w-full sm:w-auto">
          <button
            onClick={() => setIsTaskModalOpen(true)}
            className="flex-1 sm:flex-none flex items-center justify-center gap-1.5 bg-brand-600 hover:bg-brand-700 text-white font-bold text-xs px-3.5 py-2 rounded shadow-sm transition-colors"
          >
            <Plus className="w-4 h-4" />
            <span>Create Task</span>
          </button>
          <button
            onClick={() => setIsAlertModalOpen(true)}
            className="flex-1 sm:flex-none flex items-center justify-center gap-1.5 bg-red-600 hover:bg-red-700 text-white font-bold text-xs px-3.5 py-2 rounded shadow-sm transition-colors"
          >
            <AlertOctagon className="w-4 h-4" />
            <span>Broadcast Alert</span>
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* ── Kanban Board ── */}
        <div className="lg:col-span-2 space-y-4">
          <h2 className="text-base font-bold text-neutral-800 flex items-center gap-2">
            <KanbanSquare className="w-5 h-5 text-brand-500" />
            <span>Task Board — drag cards between columns</span>
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {(['todo', 'in_progress', 'done'] as const).map((col) => {
              const isOver = dragOverCol === col
              const meta   = colMeta[col]
              return (
                <div
                  id={`kanban-col-${col}`}
                  key={col}
                  onDragOver={(e) => handleDragOver(e, col)}
                  onDragLeave={handleDragLeave}
                  onDrop={(e) => handleDrop(e, col)}
                  className={`rounded-xl p-4 flex flex-col gap-3 min-h-[350px] border-2 transition-all duration-150 ${
                    isOver
                      ? 'bg-brand-50 border-brand-400 scale-[1.01]'
                      : `bg-zinc-100 ${meta.border}`
                  }`}
                >
                  <span className="text-xs uppercase font-mono font-bold text-zinc-500 tracking-wider">
                    {meta.label} ({getColTasks(col).length})
                  </span>

                  <div className="space-y-3 overflow-y-auto max-h-[450px] pr-1">
                    {getColTasks(col).map((t) => (
                      <div
                        key={t.id}
                        draggable
                        onDragStart={(e) => handleDragStart(e, t.id)}
                        onDragEnd={handleDragEnd}
                        onTouchStart={(e) => handleTouchStart(e, t.id)}
                        onTouchMove={handleTouchMove}
                        onTouchEnd={handleTouchEnd}
                        className="bg-white border border-zinc-200 rounded-lg p-3.5 shadow-sm hover:shadow-md transition-all cursor-grab active:cursor-grabbing group hover:border-brand-300 select-none"
                      >
                        <div className="flex items-start gap-2">
                          {/* Drag handle dots */}
                          <div className="mt-0.5 flex flex-col gap-[3px] flex-shrink-0 opacity-25 group-hover:opacity-60 transition-opacity">
                            <span className="block w-3.5 h-[2px] rounded-full bg-zinc-400" />
                            <span className="block w-3.5 h-[2px] rounded-full bg-zinc-400" />
                            <span className="block w-3.5 h-[2px] rounded-full bg-zinc-400" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <h4 className="font-bold text-xs text-neutral-800 leading-snug group-hover:text-brand-600">
                              {t.title}
                            </h4>
                            <p className="text-[10px] text-zinc-500 mt-1 line-clamp-2 leading-relaxed">
                              {t.description}
                            </p>
                            <div className="flex items-center justify-between mt-3">
                              <span className="text-[9px] font-mono text-zinc-400">@{t.assignedTo}</span>
                              <Badge variant={t.priority === 'high' ? 'error' : t.priority === 'medium' ? 'warning' : 'info'}>
                                {t.priority}
                              </Badge>
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}

                    {/* Empty drop target */}
                    {getColTasks(col).length === 0 && (
                      <div className={`flex items-center justify-center h-24 rounded-lg border-2 border-dashed transition-colors ${
                        isOver ? 'border-brand-400 text-brand-500' : 'border-zinc-300 text-zinc-400'
                      }`}>
                        <span className="text-[10px] font-medium">Drop here</span>
                      </div>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        </div>

        {/* ── Medical / Emergency column ── */}
        <div className="space-y-6">
          <h2 className="text-base font-bold text-neutral-800 flex items-center gap-2">
            <HeartHandshake className="w-5 h-5 text-red-500" />
            <span>First Aid &amp; Medical Dispatches</span>
          </h2>

          <div className="space-y-4 max-h-[500px] overflow-y-auto pr-1">
            {medicalReqs.length > 0 ? (
              medicalReqs.map((m) => (
                <div key={m.id} className="bg-white border border-zinc-200 rounded-xl p-4 shadow-sm space-y-3">
                  <div className="flex justify-between items-start">
                    <div>
                      <h4 className="font-bold text-xs text-neutral-800">Medical Request: {m.location}</h4>
                      <span className="text-[9px] text-zinc-400 font-mono">ID: {m.id}</span>
                    </div>
                    <Badge variant={m.status === 'resolved' ? 'success' : m.status === 'assigned' ? 'warning' : 'error'}>
                      {m.status.toUpperCase()}
                    </Badge>
                  </div>

                  <p className="text-xs text-zinc-500 leading-relaxed">{m.description}</p>

                  <div className="flex gap-2 pt-1">
                    {m.status === 'pending' && (
                      <button
                        onClick={() => handleMedicalAction(m.id, 'assign')}
                        className="w-full py-1.5 px-3 border border-zinc-300 text-xs font-semibold rounded bg-zinc-50 hover:bg-neutral-100 text-neutral-800 transition-colors"
                      >
                        Accept &amp; Dispatch
                      </button>
                    )}
                    {m.status === 'assigned' && (
                      <button
                        onClick={() => handleMedicalAction(m.id, 'resolve')}
                        className="w-full py-1.5 px-3 border border-brand-200 text-xs font-bold rounded bg-brand-50 hover:bg-brand-100 text-brand-700 flex items-center justify-center gap-1 transition-colors"
                      >
                        <Check className="w-3.5 h-3.5" />
                        <span>Mark Resolved</span>
                      </button>
                    )}
                  </div>
                </div>
              ))
            ) : (
              <div className="bg-white border border-zinc-200 rounded-xl p-6 text-center text-zinc-400">
                <p className="text-xs">No active medical dispatch requests.</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Task Creation Modal */}
      <Modal isOpen={isTaskModalOpen} onClose={() => setIsTaskModalOpen(false)} title="Create New Task">
        <form onSubmit={handleCreateTask} className="space-y-4 text-sm">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-zinc-500">Task Title</label>
            <input
              type="text"
              value={taskTitle}
              onChange={(e) => setTaskTitle(e.target.value)}
              className="w-full border border-zinc-300 rounded p-2 text-sm focus:outline-none focus:border-brand-500"
              placeholder="e.g. Turnstile scanner troubleshooting"
              required
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-zinc-500">Description</label>
            <textarea
              value={taskDesc}
              onChange={(e) => setTaskDesc(e.target.value)}
              className="w-full border border-zinc-300 rounded p-2 text-sm focus:outline-none focus:border-brand-500 h-20 resize-none"
              placeholder="Detailed explanation of requirements..."
              required
            />
          </div>
          <CustomSelect
            label="Assign To Staff"
            value={taskAssigned}
            onChange={setTaskAssigned}
            options={staffOptions}
          />
          <div className="grid grid-cols-2 gap-4">
            <CustomSelect
              label="Department"
              value={taskDept}
              onChange={setTaskDept}
              options={deptOptions}
            />
            <CustomSelect
              label="Priority"
              value={taskPriority}
              onChange={setTaskPriority}
              options={priorityOptions}
            />
          </div>
          <button
            type="submit"
            className="w-full bg-brand-600 hover:bg-brand-700 text-white font-bold py-2 rounded transition-colors text-xs"
          >
            Dispatch Task
          </button>
        </form>
      </Modal>

      {/* Alert Creation Modal */}
      <Modal isOpen={isAlertModalOpen} onClose={() => setIsAlertModalOpen(false)} title="Broadcast Emergency Warning">
        <form onSubmit={handleCreateAlert} className="space-y-4 text-sm">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-zinc-500">Alert Title</label>
            <input
              type="text"
              value={alertTitle}
              onChange={(e) => setAlertTitle(e.target.value)}
              className="w-full border border-zinc-300 rounded p-2 text-sm focus:outline-none focus:border-red-500"
              placeholder="e.g. Stand Gate A Closed"
              required
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-zinc-500">Alert Details</label>
            <textarea
              value={alertContent}
              onChange={(e) => setAlertContent(e.target.value)}
              className="w-full border border-zinc-300 rounded p-2 text-sm focus:outline-none focus:border-red-500 h-20 resize-none"
              placeholder="Instruction details for visitors..."
              required
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <CustomSelect
              label="Alert Type"
              value={alertType}
              onChange={setAlertType}
              options={alertTypeOptions}
            />
            <CustomSelect
              label="Severity"
              value={alertSeverity}
              onChange={setAlertSeverity}
              options={alertSeverityOptions}
            />
          </div>
          <button
            type="submit"
            className="w-full bg-red-600 hover:bg-red-700 text-white font-bold py-2 rounded transition-colors text-xs flex items-center justify-center gap-1.5"
          >
            <ShieldAlert className="w-4 h-4" />
            <span>Broadcast Warning Wide</span>
          </button>
        </form>
      </Modal>
    </div>
  )
}
