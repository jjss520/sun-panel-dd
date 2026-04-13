<template>
  <div class="flex justify-center items-center min-h-screen bg-slate-900">
    <NCard class="card login-card" bordered>
      <div class="card-title">
        <div class="text-center text-2xl mb-1">
          <NGradientText :gradient="{ deg: 90, from: '#3b82f6', to: '#8b5cf6' }">
            Sun Panel V2
          </NGradientText>
        </div>
        <div class="text-center text-sm text-slate-400">
          {{ $t('register.title') }}
        </div>
      </div>
      
      <NForm :model="form" :rules="rules" ref="formRef" class="mt-6">
        <NFormItem path="username">
          <NInput
            v-model:value="form.username"
            placeholder="请输入用户名"
            type="text"
          >
            <template #prefix>
              <SvgIcon icon="user" class="w-4 h-4 text-slate-500" />
            </template>
          </NInput>
        </NFormItem>
        
        <NFormItem path="password">
          <NInput
            v-model:value="form.password"
            placeholder="请输入密码"
            type="password"
            show-password-on="click"
          >
            <template #prefix>
              <SvgIcon icon="lock" class="w-4 h-4 text-slate-500" />
            </template>
          </NInput>
        </NFormItem>
        
        <NFormItem path="confirmPassword">
          <NInput
            v-model:value="form.confirmPassword"
            placeholder="请确认密码"
            type="password"
            show-password-on="click"
          >
            <template #prefix>
              <SvgIcon icon="check" class="w-4 h-4 text-slate-500" />
            </template>
          </NInput>
        </NFormItem>
        
        <NFormItem path="email">
          <NInput
            v-model:value="form.email"
            placeholder="请输入邮箱"
            type="text"
          >
            <template #prefix>
              <SvgIcon icon="email" class="w-4 h-4 text-slate-500" />
            </template>
          </NInput>
        </NFormItem>
        
        <NFormItem>
          <NButton type="primary" block :loading="loading" @click="handleRegister">
            {{ $t('register.registerButton') }}
          </NButton>
        </NFormItem>
        
        <NFormItem>
          <NButton type="default" block @click="() => router.push('/login')">
            {{ $t('register.backToLogin') }}
          </NButton>
        </NFormItem>
      </NForm>
      
      <div class="flex justify-center text-slate-300 mt-5">
        Powered By <a href="https://github.com/jjss520/sun-panel-dd" target="_blank" class="ml-[5px] text-slate-500">Sun-Panel-V2</a>
      </div>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import { NButton, NCard, NForm, NFormItem, NGradientText, NInput, useMessage } from 'naive-ui'
import { ref } from 'vue'
import { commit as register } from '@/api/register'
import { router } from '@/router'
import { t } from '@/locales'
import { SvgIcon } from '@/components/common'
import type { FormInst } from 'naive-ui'

const formRef = ref<FormInst | null>(null)
const message = useMessage()
const loading = ref(false)

const form = ref<{
  username: string
  password: string
  confirmPassword: string
  email: string
}>({
  username: '',
  password: '',
  confirmPassword: '',
  email: ''
})

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 30, message: '密码长度在 6 到 30 个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator(rule: any, value: string) {
        if (!value || form.value.password === value) {
          return true
        }
        return new Error('两次输入的密码不一致')
      },
      trigger: 'blur'
    }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    {
      pattern: /^[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)+$/,
      message: '请输入正确的邮箱地址',
      trigger: 'blur'
    }
  ]
}

// 移除了JSX的renderPrefix函数，改为使用Vue的模板语法

// 注册次数限制配置
const REGISTER_LIMIT = {
  MAX_ATTEMPTS: 5, // 最大尝试次数
  TIME_WINDOW: 3600 * 1000, // 时间窗口（1小时）
  BLOCK_DURATION: 3600 * 1000, // 超出限制后的阻止时间（1小时）
}

// 获取注册尝试记录
const getRegistrationAttempts = () => {
  const attempts = localStorage.getItem('registration_attempts')
  if (attempts) {
    return JSON.parse(attempts)
  }
  return {
    count: 0,
    timestamp: Date.now(),
    blockedUntil: 0
  }
}

// 更新注册尝试记录
const updateRegistrationAttempts = (attempts: any) => {
  localStorage.setItem('registration_attempts', JSON.stringify(attempts))
}

// 检查是否达到注册限制
const checkRegistrationLimit = () => {
  const attempts = getRegistrationAttempts()
  const now = Date.now()
  
  // 如果在阻止时间内，返回错误信息
  if (attempts.blockedUntil > now) {
    const remainingTime = Math.ceil((attempts.blockedUntil - now) / 60000) // 剩余时间（分钟）
    return {
      exceeded: true,
      message: `注册尝试次数过多，请 ${remainingTime} 分钟后再试`
    }
  }
  
  // 如果超过时间窗口，重置计数
  if (now - attempts.timestamp > REGISTER_LIMIT.TIME_WINDOW) {
    updateRegistrationAttempts({
      count: 0,
      timestamp: now,
      blockedUntil: 0
    })
    return { exceeded: false }
  }
  
  // 如果超过最大尝试次数，设置阻止时间
  if (attempts.count >= REGISTER_LIMIT.MAX_ATTEMPTS) {
    attempts.blockedUntil = now + REGISTER_LIMIT.BLOCK_DURATION
    updateRegistrationAttempts(attempts)
    return {
      exceeded: true,
      message: `注册尝试次数过多，请 ${REGISTER_LIMIT.BLOCK_DURATION / 60000} 分钟后再试`
    }
  }
  
  return { exceeded: false }
}

// 更新注册尝试计数
const updateAttemptCount = () => {
  const attempts = getRegistrationAttempts()
  const now = Date.now()
  
  // 如果超过时间窗口，重置计数
  if (now - attempts.timestamp > REGISTER_LIMIT.TIME_WINDOW) {
    updateRegistrationAttempts({
      count: 1,
      timestamp: now,
      blockedUntil: 0
    })
  } else {
    updateRegistrationAttempts({
      ...attempts,
      count: attempts.count + 1
    })
  }
}

// 重置注册尝试计数（注册成功后）
const resetAttemptCount = () => {
  updateRegistrationAttempts({
    count: 0,
    timestamp: Date.now(),
    blockedUntil: 0
  })
}

const handleRegister = async () => {
  if (!formRef.value) return
  
  // 检查注册次数限制
  const limitCheck = checkRegistrationLimit()
  if (limitCheck.exceeded && limitCheck.message) {
      message.error(limitCheck.message)
      return
    }
  
  try {
    await formRef.value.validate()
    loading.value = true
    
    const response = await register({
      username: form.value.username,
      password: form.value.password,
      email: form.value.email
    })
    
    if (response.code === 0) {
      message.success(t('register.registerSuccess'))
      // 注册成功后重置尝试计数
      resetAttemptCount()
      // 跳转到登录页面
      setTimeout(() => {
        router.push('/login')
      }, 1500)
    } else {
      message.error(response.msg || t('register.registerFailed'))
      // 注册失败，更新尝试计数
      updateAttemptCount()
    }
  } catch (error: any) {
    console.error('注册失败:', error)
    message.error(error?.message || t('register.registerFailed'))
    // 注册失败，更新尝试计数
    updateAttemptCount()
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.card {
  width: 420px;
  max-width: 100%;
  padding: 24px;
  margin: 20px;
}

.login-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(71, 85, 105, 0.3);
}

.card-title {
  margin-bottom: 20px;
}
</style>