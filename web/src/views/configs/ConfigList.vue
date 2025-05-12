<template>
  <div class="config-list">
    <div class="page-header">
      <h2>配置管理</h2>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        添加配置
      </el-button>
    </div>

    <el-card class="config-table">
      <el-table
        v-loading="loading"
        :data="configList"
        border
        style="width: 100%"
      >
        <el-table-column prop="key" label="配置键" min-width="200">
          <template #default="{ row }">
            <router-link :to="'/configs/' + row.key" class="config-key">
              {{ row.key }}
            </router-link>
          </template>
        </el-table-column>
        <el-table-column prop="value" label="配置值" min-width="300">
          <template #default="{ row }">
            <el-tag
              :type="getValueType(row.value).type"
              effect="plain"
              size="small"
            >
              {{ getValueType(row.value).label }}
            </el-tag>
            {{ formatValue(row.value) }}
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column prop="updatedBy" label="更新人" width="120" />
        <el-table-column prop="updatedAt" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.updatedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button
                type="primary"
                link
                @click="handleEdit(row)"
              >
                编辑
              </el-button>
              <el-button
                type="primary"
                link
                @click="handleHistory(row)"
              >
                历史
              </el-button>
              <el-button
                type="danger"
                link
                @click="handleDelete(row)"
              >
                删除
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 配置编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '添加配置' : '编辑配置'"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="配置键" prop="key">
          <el-input
            v-model="form.key"
            :disabled="dialogType === 'edit'"
            placeholder="请输入配置键"
          />
        </el-form-item>
        <el-form-item label="值类型" prop="valueType">
          <el-select v-model="form.valueType" placeholder="请选择值类型">
            <el-option label="字符串" value="string" />
            <el-option label="数字" value="number" />
            <el-option label="布尔值" value="boolean" />
            <el-option label="JSON" value="json" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置值" prop="value">
          <el-input
            v-if="form.valueType === 'string'"
            v-model="form.value"
            placeholder="请输入字符串值"
          />
          <el-input-number
            v-else-if="form.valueType === 'number'"
            v-model="form.value"
            :precision="2"
            :step="0.1"
            placeholder="请输入数字值"
          />
          <el-switch
            v-else-if="form.valueType === 'boolean'"
            v-model="form.value"
          />
          <el-input
            v-else
            v-model="form.value"
            type="textarea"
            :rows="4"
            placeholder="请输入JSON格式的值"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 历史版本对话框 -->
    <el-dialog
      v-model="historyDialogVisible"
      title="配置历史"
      width="800px"
    >
      <el-timeline>
        <el-timeline-item
          v-for="item in historyList"
          :key="item.version"
          :timestamp="formatTime(item.updatedAt)"
          :type="item.version === currentVersion ? 'primary' : ''"
        >
          <h4>版本 {{ item.version }}</h4>
          <p>更新人：{{ item.updatedBy }}</p>
          <p>配置值：{{ formatValue(item.value) }}</p>
        </el-timeline-item>
      </el-timeline>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import { useConfigStore } from '@/stores/config'

const configStore = useConfigStore()
const loading = ref(false)
const configList = ref([])
const dialogVisible = ref(false)
const dialogType = ref('add')
const historyDialogVisible = ref(false)
const historyList = ref([])
const currentVersion = ref(0)

const form = ref({
  key: '',
  value: '',
  valueType: 'string'
})

const rules = {
  key: [
    { required: true, message: '请输入配置键', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_.]*$/, message: '配置键格式不正确', trigger: 'blur' }
  ],
  valueType: [
    { required: true, message: '请选择值类型', trigger: 'change' }
  ],
  value: [
    { required: true, message: '请输入配置值', trigger: 'blur' }
  ]
}

const formRef = ref(null)

// 加载配置列表
const loadConfigList = async () => {
  loading.value = true
  try {
    const list = await configStore.getConfigList()
    configList.value = list
  } catch (error) {
    ElMessage.error('加载配置列表失败')
  } finally {
    loading.value = false
  }
}

// 获取值类型
const getValueType = (value) => {
  const type = typeof value
  switch (type) {
    case 'string':
      return { type: '', label: '字符串' }
    case 'number':
      return { type: 'success', label: '数字' }
    case 'boolean':
      return { type: 'warning', label: '布尔值' }
    case 'object':
      return { type: 'info', label: 'JSON' }
    default:
      return { type: 'danger', label: '未知' }
  }
}

// 格式化值
const formatValue = (value) => {
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value)
}

// 格式化时间
const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 添加配置
const handleAdd = () => {
  dialogType.value = 'add'
  form.value = {
    key: '',
    value: '',
    valueType: 'string'
  }
  dialogVisible.value = true
}

// 编辑配置
const handleEdit = (row) => {
  dialogType.value = 'edit'
  form.value = {
    key: row.key,
    value: row.value,
    valueType: typeof row.value
  }
  dialogVisible.value = true
}

// 查看历史版本
const handleHistory = async (row) => {
  try {
    const history = await configStore.getConfigHistory(row.key)
    historyList.value = history
    currentVersion.value = row.version
    historyDialogVisible.value = true
  } catch (error) {
    ElMessage.error('加载配置历史失败')
  }
}

// 删除配置
const handleDelete = (row) => {
  ElMessageBox.confirm(
    '确定要删除该配置吗？',
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await configStore.deleteConfig(row.key)
      ElMessage.success('删除成功')
      loadConfigList()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const value = parseConfigValue(form.value.value, form.value.valueType)
        if (dialogType.value === 'add') {
          await configStore.addConfig(form.value.key, value)
          ElMessage.success('添加成功')
        } else {
          await configStore.updateConfig(form.value.key, value)
          ElMessage.success('更新成功')
        }
        dialogVisible.value = false
        loadConfigList()
      } catch (error) {
        ElMessage.error(dialogType.value === 'add' ? '添加失败' : '更新失败')
      }
    }
  })
}

// 解析配置值
const parseConfigValue = (value, type) => {
  switch (type) {
    case 'number':
      return Number(value)
    case 'boolean':
      return Boolean(value)
    case 'json':
      try {
        return JSON.parse(value)
      } catch (error) {
        throw new Error('JSON格式不正确')
      }
    default:
      return value
  }
}

onMounted(() => {
  loadConfigList()
})
</script>

<style lang="scss" scoped>
.config-list {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    
    h2 {
      margin: 0;
      font-weight: 500;
    }
  }
  
  .config-table {
    .config-key {
      color: #409eff;
      text-decoration: none;
      
      &:hover {
        text-decoration: underline;
      }
    }
  }
}
</style> 